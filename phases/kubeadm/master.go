package kubeadm

import (
	"fmt"
	cmdtext "github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/rundata"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog"
	"net"
	"strings"
)

// InitMaster init master0
func InitMaster(node *rundata.Node, kubeadmCfg *rundata.Kubeadm) error {
	apiDomainName, _, _ := net.SplitHostPort(kubeadmCfg.ControlPlaneEndpoint)
	klog.Infof("[%s] [kubeadm] Initializing master0", node.HostInfo.Host)

	if err := setHost(node, "127.0.0.1", apiDomainName); err != nil {
		return fmt.Errorf("[%s] Failed to set /etc/hosts: %v", node.HostInfo.Host, err)
	}

	output, err := initMaster(node, kubeadmCfg)
	if err != nil {
		return fmt.Errorf("[%s] Failed to Initialize master0", node.HostInfo.Host, err)
	}
	klog.Infof("[%s] [kubeadm] Successfully initialized master0", node.HostInfo.Host)

	if err := copyAdminConfig(node); err != nil {
		return fmt.Errorf("[%s] [config] Failed to copy admin.conf to $HOME/.kube/config: %v", node.HostInfo.Host, err)
	}

	klog.V(2).Infof("[%s] [token] Getting token from master init output", node.HostInfo.Host)
	setToken(string(output), &kubeadmCfg.Token)

	return nil

}

func initMaster(node *rundata.Node, kubeadmCfg *rundata.Kubeadm) ([]byte, error) {
	text, err := cmdtext.Kubeadm(cmdtext.Init, node.Name, kubeadmCfg)
	if err != nil {
		return nil, err
	}

	output, err := node.SSH.RunOut(text)
	if err != nil {
		return nil, err
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
			klog.Infof("[%s] [kubeadm] Joining master nodes", node.HostInfo.Host)
			if err := setHost(node, masters[0].HostInfo.Host, apiDomainName); err != nil {
				return fmt.Errorf("[%s] Failed to set /etc/hosts: %v", node.HostInfo.Host, err)
			}

			if err := joinControlPlane(node, kubeadmCfg); err != nil {
				return fmt.Errorf("[%s] Failed to join master nodes: %v", node.HostInfo.Host, err)
			}

			if err := copyAdminConfig(node); err != nil {
				return fmt.Errorf("[%s] [config] Failed to copy admin.conf to $HOME/.kube/config: %v", node.HostInfo.Host, err)
			}

			if err := setHost(node, "127.0.0.1", apiDomainName); err != nil {
				return fmt.Errorf("[%s] Failed to set /etc/hosts: %v", node.HostInfo.Host, err)
			}
			klog.Infof("[%s] [kubeadm] Successfully joined master nodes", node.HostInfo.Host)

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
		return err
	}

	if err := node.SSH.Run(text); err != nil {
		return err
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

func setHost(node *rundata.Node, ip, apiDomainName string) error {
	klog.V(2).Infof("[%s] [host] Add \"%s %s\" to /etc/hosts", node.HostInfo.Host, ip, apiDomainName)
	if err := node.SSH.Run(cmdtext.SetHosts(ip, apiDomainName)); err != nil {
		return err
	}
	return nil
}

func changeHost(node *rundata.Node, oldIP, newIP, apiDomainName string) error {
	klog.V(2).Infof("[%s] [host] Change \"%s %s\" to \"%s %s\" on /etc/hosts", node.HostInfo.Host, oldIP, apiDomainName, newIP, apiDomainName)
	if err := node.SSH.Run(cmdtext.ChangeHosts(newIP, apiDomainName)); err != nil {
		return err
	}
	return nil
}

func copyAdminConfig(node *rundata.Node) error {
	klog.V(2).Info("[%s] [config] Copy admin.conf to $HOME/.kube/config")
	if err := node.SSH.Run(cmdtext.CopyAdminConfig()); err != nil {
		return err
	}
	return nil
}
