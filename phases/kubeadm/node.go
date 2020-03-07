package kubeadm

import (
	"context"
	"fmt"
	"github.com/bilibili/kratos/pkg/sync/errgroup"
	cmdtext "github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"net"
	"strings"
	"time"
)

//JoinNode join nodes
func JoinNode(cfg *rundata.Kubei, kubeadmCfg *rundata.Kubeadm) error {

	apiDomainName, _, _ := net.SplitHostPort(kubeadmCfg.ControlPlaneEndpoint)
	var mastersIP []string
	for _, master := range cfg.ClusterNodes.Masters {
		mastersIP = append(mastersIP, master.HostInfo.Host)
	}

	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(20)
	for _, node := range cfg.ClusterNodes.Worker {
		node := node
		g.Go(func(ctx context.Context) error {

			if err := system.SwapOff(node); err != nil {
				return err
			}

			if err := iptables(node); err != nil {
				return err
			}

			if cfg.IsHA {
				if err := system.SetHost(node, constants.LoopbackAddress, apiDomainName); err != nil {
					return err
				}

				klog.Infof("[%s] [slb] Setting up the local SLB", node.HostInfo.Host)
				if err := localSLB(mastersIP, node, kubeadmCfg); err != nil {
					return fmt.Errorf("[%s] Failed to set up the local SLB: %v", node.HostInfo.Host, err)
				}

				klog.Infof("[%s] [slb] Successfully set up the local SLB", node.HostInfo.Host)
			} else {
				// set /etc/hosts
				if err := system.SetHost(node, mastersIP[0], apiDomainName); err != nil {
					return err
				}

			}

			// join worker node
			klog.Infof("[%s] [kubeadm] Joining worker nodes", node.HostInfo.Host)
			if err := joinNode(node, kubeadmCfg); err != nil {
				return fmt.Errorf("[%s] Failed to join master worker : %v", node.HostInfo.Host, err)
			}
			klog.Infof("[%s] [kubeadm] Successfully joined worker nodes", node.HostInfo.Host)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func joinNode(node *rundata.Node, kubeadmCfg *rundata.Kubeadm) error {
	text, err := cmdtext.Kubeadm(cmdtext.JoinNode, node.Name, kubeadmCfg)
	if err != nil {
		return err
	}
	if err := node.SSH.Run(text); err != nil {
		return err
	}
	return nil
}

func localSLB(masters []string, node *rundata.Node, kubeadmCfg *rundata.Kubeadm) error {
	text, err := cmdtext.NginxConf(masters, "6443")
	if err != nil {
		return err
	}
	if err := node.SSH.Run(text); err != nil {
		return err
	}

	if err := node.SSH.Run(cmdtext.NginxManifest("nginx:1.17")); err != nil {
		return err
	}

	if err := node.SSH.Run(cmdtext.KubeletUnitFile(fmt.Sprintf("%s/%s", kubeadmCfg.ImageRepository, "pause:3.1"))); err != nil {
		return err
	}

	klog.V(2).Infof("[%s] [restart] restart kubelet to boot up the nginx proxy as static Pod", node.HostInfo.Host)

	if err := system.Restart("kubelet", node); err != nil {
		return err
	}

	klog.V(2).Infof("[%s] [slb] Waiting for the kubelet to boot up the nginx proxy as static Pod. This can take up to %v", node.HostInfo.Host, constants.DefaultLocalSLBTimeout)
	if err := checkHealth(node, fmt.Sprintf("https://%s/%s", kubeadmCfg.ControlPlaneEndpoint, "healthz"), constants.DefaultLocalSLBInterval, constants.DefaultLocalSLBTimeout); err != nil {
		return err
	}

	if err := node.SSH.Run(cmdtext.RemoveKubeletUnitFile()); err != nil {
		return err
	}

	if err := system.Restart("kubelet", node); err != nil {
		return err
	}

	return nil
}

func checkHealth(node *rundata.Node, url string, interval, timeout time.Duration) error {

	if err := wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		var output []byte
		output, _ = node.SSH.RunOut(fmt.Sprintf("curl -k %s", url))
		if string(output) == "ok" {
			return true, nil
		}

		return false, nil
	}); err != nil {
		return err
	}

	return nil
}

func iptables(node *rundata.Node) error {
	klog.V(2).Infof("[%s] [iptables] set up iptables", node.HostInfo.Host)
	if err := node.SSH.Run(cmdtext.Iptables()); err != nil {
		return fmt.Errorf("[%s] [iptables] Failed set up iptables: %v", node.HostInfo.Host, err)
	}
	return nil
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
