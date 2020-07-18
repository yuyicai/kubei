package kubeadm

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/pkg/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"github.com/yuyicai/kubei/tmpl"
)

//JoinNode join nodes
func JoinNode(c *rundata.Cluster) error {
	return c.RunOnWorkers(func(node *rundata.Node) error {
		if err := system.SwapOff(node); err != nil {
			return err
		}

		if err := iptables(node); err != nil {
			return err
		}

		if err := ha(node, c.Kubei.ClusterNodes.GetAllMastersHost(), &c.Kubei.HA, c.Kubeadm); err != nil {
			return err
		}

		// join worker node
		klog.Infof("[%s] [kubeadm-join] Joining worker nodes", node.HostInfo.Host)
		if err := joinNode(node, *c.Kubei, *c.Kubeadm); err != nil {
			return fmt.Errorf("[%s] Failed to join master worker : %v", node.HostInfo.Host, err)
		}
		klog.Infof("[%s] [kubeadm-join] Successfully joined worker nodes", node.HostInfo.Host)

		return nil
	})
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
	if err := node.Run(text); err != nil {
		return err
	}

	if err := node.Run(tmpl.NginxManifest(n.Image.GetImage())); err != nil {
		return err
	}

	if err := node.Run(tmpl.KubeletUnitFile(fmt.Sprintf("%s/%s", kcfg.ImageRepository, "pause:3.1"))); err != nil {
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

	if err := node.Run(tmpl.RemoveKubeletUnitFile()); err != nil {
		return err
	}

	return system.Restart("kubelet", node)
}

func checkHealth(node *rundata.Node, url string, interval, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		var output []byte
		output, _ = node.RunOut(fmt.Sprintf("curl -k %s", url))
		if string(output) == "ok" {
			return true, nil
		}

		return false, nil
	})
}

func iptables(node *rundata.Node) error {
	klog.V(2).Infof("[%s] [iptables] set up iptables", node.HostInfo.Host)
	if err := node.Run(tmpl.Iptables()); err != nil {
		return fmt.Errorf("[%s] [iptables] Failed set up iptables: %v", node.HostInfo.Host, err)
	}
	return nil
}

func joinNode(node *rundata.Node, kubeiCfg rundata.Kubei, kubeadmCfg rundata.Kubeadm) error {
	text, err := tmpl.Kubeadm(tmpl.JoinNode, node.Name, kubeiCfg.Kubernetes, kubeadmCfg)
	if err != nil {
		return err
	}
	return node.Run(text)
}

func CheckNodesReady(c *rundata.Cluster) error {
	return c.RunOnFirstMaster(func(node *rundata.Node) error {
		output, err := checkNodesReady(node, constants.DefaultWaitNodeInterval, constants.DefaultWaitNodeTimeout)
		if err != nil {
			return err
		}
		klog.Info(output, "\nKubernetes High-Availability cluster deployment completed")
		return nil
	})
}

func checkNodesReady(node *rundata.Node, interval, timeout time.Duration) (string, error) {
	var str string
	klog.Infof("[%s] [check] Waiting for all nodes to become ready. This can take up to %v", node.HostInfo.Host, timeout)
	if err := wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		var output []byte
		output, _ = node.RunOut("kubectl get nodes -owide")
		str = string(output)
		if strings.Contains(str, "NotReady") {
			return false, nil
		} else if strings.Contains(str, "Ready") {
			return true, nil
		}
		return false, nil
	}); err != nil {
		return "", err
	}
	return str, nil
}

func LoadOfflineImages(c *rundata.Cluster) error {

	g := errgroup.WithCancel(context.Background())
	g.Go(func(ctx context.Context) error {
		if err := c.RunOnMasters(func(node *rundata.Node) error {
			return loadOfflineImagesOnnode("master", node)
		}); err != nil {
			return err
		}

		if err := c.RunOnAllNodes(func(node *rundata.Node) error {
			return loadOfflineImagesOnnode("node", node)
		}); err != nil {
			return err
		}
		return nil
	})

	return g.Wait()
}

func loadOfflineImagesOnnode(nodeType string, node *rundata.Node) error {
	if node.InstallType == constants.InstallTypeOffline {
		return node.Run(fmt.Sprintf("sh /tmp/.kubei/images/%s.sh", nodeType))
	}
	return nil
}
