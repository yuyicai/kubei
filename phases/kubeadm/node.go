package kubeadm

import (
	"context"
	"fmt"
	"github.com/bilibili/kratos/pkg/sync/errgroup"
	"github.com/yuyicai/kubei/cmd/tmpl"
	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"net"
	"strconv"
	"strings"
	"time"
)

//JoinNode join nodes
func JoinNode(kubeiCfg *rundata.Kubei, kubeadmCfg *rundata.Kubeadm) error {

	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range kubeiCfg.ClusterNodes.Worker {
		node := node
		g.Go(func(ctx context.Context) error {

			if err := system.SwapOff(node); err != nil {
				return err
			}

			if err := iptables(node); err != nil {
				return err
			}

			if err := ha(node, kubeiCfg.ClusterNodes.GetAllMastersHost(), &kubeiCfg.HA, kubeadmCfg); err != nil {
				return err
			}

			// join worker node
			klog.Infof("[%s] [kubeadm] Joining worker nodes", node.HostInfo.Host)
			if err := joinNode(node, *kubeiCfg, *kubeadmCfg); err != nil {
				return fmt.Errorf("[%s] Failed to join master worker : %v", node.HostInfo.Host, err)
			}
			klog.Infof("[%s] [kubeadm] Successfully joined worker nodes", node.HostInfo.Host)

			return nil
		})
	}

	return g.Wait()
}

func ha(node *rundata.Node, masters []string, h *rundata.HA, kcfg *rundata.Kubeadm) error {
	apiDomainName, _, _ := net.SplitHostPort(kcfg.ControlPlaneEndpoint)

	switch h.Type {
	case constants.HATypeNone:
		return system.SetHost(node, masters[0], apiDomainName)
	case constants.HATypeLocalSLB:
		if err := system.SetHost(node, constants.LoopbackAddress, apiDomainName); err != nil {
			return err
		}

		klog.Infof("[%s] [slb] Setting up the local SLB", node.HostInfo.Host)
		if err := localSLB(masters, node, &h.LocalSLB, kcfg); err != nil {
			return fmt.Errorf("[%s] Failed to set up the local SLB: %v", node.HostInfo.Host, err)
		}
		klog.Infof("[%s] [slb] Successfully set up the local SLB", node.HostInfo.Host)
	case constants.HATypeExternalSLB:
		//TODO

	}

	return nil
}

func localSLB(masters []string, node *rundata.Node, slb *rundata.LocalSLB, kubeadmCfg *rundata.Kubeadm) error {
	switch slb.Type {
	case constants.LocalSLBTypeNginx:
		return nginx(node, &slb.Nginx, masters, kubeadmCfg)
	case constants.LocalSLBTypeHAproxy:
		//TODO
	}
	return nil
}

func nginx(node *rundata.Node, n *rundata.Nginx, masters []string, kcfg *rundata.Kubeadm) error {
	text, err := tmpl.NginxConf(masters, n.Port, strconv.FormatInt(int64(kcfg.LocalAPIEndpoint.BindPort), 10))
	if err != nil {
		return err
	}
	if err := node.SSH.Run(text); err != nil {
		return err
	}

	if err := node.SSH.Run(tmpl.NginxManifest(n.Image.GetImage())); err != nil {
		return err
	}

	if err := node.SSH.Run(tmpl.KubeletUnitFile(fmt.Sprintf("%s/%s", kcfg.ImageRepository, "pause:3.1"))); err != nil {
		return err
	}

	klog.V(2).Infof("[%s] [restart] restart kubelet to boot up the nginx proxy as static Pod", node.HostInfo.Host)

	if err := system.Restart("kubelet", node); err != nil {
		return err
	}

	klog.V(2).Infof("[%s] [slb] Waiting for the kubelet to boot up the nginx proxy as static Pod. This can take up to %v", node.HostInfo.Host, constants.DefaultLocalSLBTimeout)
	if err := checkHealth(node, fmt.Sprintf("https://%s/%s", kcfg.ControlPlaneEndpoint, "healthz"), constants.DefaultLocalSLBInterval, constants.DefaultLocalSLBTimeout); err != nil {
		return err
	}

	if err := node.SSH.Run(tmpl.RemoveKubeletUnitFile()); err != nil {
		return err
	}

	return system.Restart("kubelet", node)
}

func checkHealth(node *rundata.Node, url string, interval, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		var output []byte
		output, _ = node.SSH.RunOut(fmt.Sprintf("curl -k %s", url))
		if string(output) == "ok" {
			return true, nil
		}

		return false, nil
	})
}

func iptables(node *rundata.Node) error {
	klog.V(2).Infof("[%s] [iptables] set up iptables", node.HostInfo.Host)
	if err := node.SSH.Run(tmpl.Iptables()); err != nil {
		return fmt.Errorf("[%s] [iptables] Failed set up iptables: %v", node.HostInfo.Host, err)
	}
	return nil
}

func joinNode(node *rundata.Node, kubeiCfg rundata.Kubei, kubeadmCfg rundata.Kubeadm) error {
	text, err := tmpl.Kubeadm(tmpl.JoinNode, node.Name, kubeiCfg.Kubernetes, kubeadmCfg)
	if err != nil {
		return err
	}
	return node.SSH.Run(text)
}

func CheckNodesReady(node *rundata.Node, interval, timeout time.Duration) (string, bool) {
	var str string
	klog.Infof("[check] Waiting for all nodes to become ready. This can take up to %v", timeout)
	if err := wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		var output []byte
		output, _ = node.SSH.RunOut("kubectl get nodes -owide")
		str = string(output)
		if strings.Contains(str, "NotReady") {
			return false, nil
		} else if strings.Contains(str, "Ready") {
			return true, nil
		}

		return false, nil
	}); err != nil {
		return "", false
	}

	return str, true
}

func LoadOfflineImages(c rundata.ClusterNodes) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range c.Masters {
		node := node
		g.Go(func(ctx context.Context) error {
			return loadOfflineImagesOnnode("master", node)
		})
	}

	for _, node := range c.GetAllNodes() {
		node := node
		g.Go(func(ctx context.Context) error {
			return loadOfflineImagesOnnode("node", node)
		})
	}

	return g.Wait()
}

func loadOfflineImagesOnnode(nodeType string, node *rundata.Node) error {
	if node.InstallType == constants.InstallTypeOffline {
		return node.SSH.Run(fmt.Sprintf("sh /tmp/.kubei/images/%s.sh", nodeType))
	}
	return nil
}
