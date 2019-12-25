package kubeadm

import (
	"fmt"
	cmdtext "github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/rundata"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"net"
	"time"
)

//JoinNode join nodes
func JoinNode(cfg *rundata.Kubei, kubeadmCfg *rundata.Kubeadm) error {

	apiDomainName, _, _ := net.SplitHostPort(kubeadmCfg.ControlPlaneEndpoint)
	var mastersIP []string
	for _, master := range cfg.ClusterNodes.Masters {
		mastersIP = append(mastersIP, master.HostInfo.Host)
	}
	g := errgroup.Group{}
	for _, node := range cfg.ClusterNodes.Worker {
		node := node
		g.Go(func() error {
			if cfg.IsHA {
				if err := setHost(node, "127.0.0.1", apiDomainName); err != nil {
					return fmt.Errorf("[%s] Failed to set /etc/hosts: %v", node.HostInfo.Host, err)
				}

				klog.Infof("[%s] [slb] Setting up the local SLB", node.HostInfo.Host)
				if err := ha(mastersIP, node, kubeadmCfg); err != nil {
					return fmt.Errorf("[%s] Failed to set up the local SLB: ", node.HostInfo.Host, err)
				}

				klog.Infof("[%s] [slb] Successfully set up the local SLB", node.HostInfo.Host)
			} else {
				// set /etc/hosts
				if err := setHost(node, mastersIP[0], apiDomainName); err != nil {
					return fmt.Errorf("[%s] Failed to set /etc/hosts: %v", node.HostInfo.Host, err)
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

func ha(masters []string, node *rundata.Node, kubeadmCfg *rundata.Kubeadm) error {
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

	if err := checkHealth(node, fmt.Sprintf("https://%s/%s", kubeadmCfg.ControlPlaneEndpoint, "healthz"), 2*time.Second, 4*time.Minute); err != nil {
		return err
	}

	if err := node.SSH.Run(cmdtext.RemoveKubeletUnitFile()); err != nil {
		return err
	}

	return nil
}

func checkHealth(node *rundata.Node, url string, interval, timeout time.Duration) error {

	if err := wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		var output []byte
		output, _ = node.SSH.RunOut(fmt.Sprintf("curl -k %s", url))
		if string(output) != "ok" {
			return false, nil
		}

		return true, nil
	}); err != nil {
		return err
	}

	return nil
}
