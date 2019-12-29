package kubeadm

import (
	"fmt"
	cmdtext "github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog"
	"net"
	"strings"
)

// InitMaster init master0
func InitMaster(node *rundata.Node, kubeadmCfg *rundata.Kubeadm) error {
	apiDomainName, _, _ := net.SplitHostPort(kubeadmCfg.ControlPlaneEndpoint)

	klog.Infof("[%s] [kubeadm-init] Initializing master0", node.HostInfo.Host)

	if err := system.SetHost(node, "127.0.0.1", apiDomainName); err != nil {
		return err
	}

	if err := system.SwapOff(node); err != nil {
		return err
	}

	if err := iptables(node); err != nil {
		return err
	}

	output, err := initMaster(node, kubeadmCfg)
	if err != nil {
		return err
	}

	if err := copyAdminConfig(node); err != nil {
		return err
	}

	klog.Infof("[%s] [kubeadm-init] Successfully initialized master0", node.HostInfo.Host)

	klog.V(2).Infof("[%s] [token] Getting token from master init output", node.HostInfo.Host)
	setToken(string(output), &kubeadmCfg.Token)

	return nil

}

func initMaster(node *rundata.Node, kubeadmCfg *rundata.Kubeadm) ([]byte, error) {
	text, err := cmdtext.Kubeadm(cmdtext.Init, node.Name, kubeadmCfg)
	if err != nil {
		return nil, fmt.Errorf("[%s] [kubeadm-init] Failed to Initialize master0: %v", node.HostInfo.Host, err)
	}

	output, err := node.SSH.RunOut(text)
	if err != nil {
		return nil, fmt.Errorf("[%s] [kubeadm-init] Failed to Initialize master0: %v", node.HostInfo.Host, err)
	}
	return output, nil
}

// JoinControlPlane join masters to ControlPlane
func JoinControlPlane(masters []*rundata.Node, kubeadmCfg *rundata.Kubeadm) error {
	apiDomainName, _, _ := net.SplitHostPort(kubeadmCfg.ControlPlaneEndpoint)
	g := errgroup.Group{}
	for _, node := range masters[1:] {
		node := node
		g.Go(func() error {
			klog.Infof("[%s] [kubeadm-join] Joining master nodes", node.HostInfo.Host)
			if err := system.SetHost(node, masters[0].HostInfo.Host, apiDomainName); err != nil {
				return err
			}

			if err := system.SwapOff(node); err != nil {
				return err
			}

			if err := iptables(node); err != nil {
				return err
			}

			if err := joinControlPlane(node, kubeadmCfg); err != nil {
				return err
			}

			if err := copyAdminConfig(node); err != nil {
				return err
			}

			if err := system.SetHost(node, "127.0.0.1", apiDomainName); err != nil {
				return err
			}
			klog.Infof("[%s] [kubeadm-join] Successfully joined master nodes", node.HostInfo.Host)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil

}

func joinControlPlane(node *rundata.Node, kubeadmCfg *rundata.Kubeadm) error {
	text, err := cmdtext.Kubeadm(cmdtext.JoinControlPlane, node.Name, kubeadmCfg)
	if err != nil {
		return fmt.Errorf("[%s] [kubeadm-join] Failed to join master nodes: %v", node.HostInfo.Host, err)
	}

	if err := node.SSH.Run(text); err != nil {
		return fmt.Errorf("[%s] [kubeadm-join] Failed to join master nodes: %v", node.HostInfo.Host, err)
	}

	return nil
}

//setToken get token from kubeadm init output
func setToken(str string, token *rundata.Token) {
	if strSlice := strings.Split(str, "--token "); len(strSlice) > 1 {
		token.Token = strSlice[1][:23]
	}

	if strSlice := strings.Split(str, "sha256:"); len(strSlice) > 1 {
		token.CaCertHash = strSlice[1][:64]
	}

	if strSlice := strings.Split(str, "--certificate-key "); len(strSlice) > 1 {
		token.CertificateKey = strSlice[1][:64]
	}
}

func copyAdminConfig(node *rundata.Node) error {
	klog.V(2).Infof("[%s] [kubectl-config] Copy admin.conf to $HOME/.kube/config", node.HostInfo.Host)
	if err := node.SSH.Run(cmdtext.CopyAdminConfig()); err != nil {
		return fmt.Errorf("[%s] [kubectl-config] Failed to copy admin.conf to $HOME/.kube/config: %v", node.HostInfo.Host, err)
	}

	if node.HostInfo.User != "root" {
		klog.V(2).Infof("[%s] [kubectl-config] Chown $HOME/.kube/config to user %s", node.HostInfo.Host, node.HostInfo.User)
		if err := node.SSH.Run(cmdtext.ChownKubectlConfig()); err != nil {
			return fmt.Errorf("[%s] [kubectl-config] Failed to chown $HOME/.kube/config to user %s: %v", node.HostInfo.Host, node.HostInfo.User, err)
		}
	}

	return nil
}
