package kubeadm

import (
	"fmt"
	"net"
	"strings"

	"github.com/fatih/color"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/internal/constants"
	"github.com/yuyicai/kubei/internal/phases/system"
	"github.com/yuyicai/kubei/internal/rundata"
	"github.com/yuyicai/kubei/internal/tmpl"
)

// InitMaster init master0
func InitMaster(c *rundata.Cluster) error {
	color.HiBlue("Initializing master0 ☸️")
	return c.RunOnFirstMaster(func(node *rundata.Node, c *rundata.Cluster) error {
		apiDomainName, _, _ := net.SplitHostPort(c.Kubeadm.ControlPlaneEndpoint)
		if err := system.SetHost(node, constants.LoopbackAddress, apiDomainName); err != nil {
			return err
		}

		if err := system.SwapOff(node); err != nil {
			return err
		}

		if err := iptables(node); err != nil {
			return err
		}

		klog.V(3).Infof("[%s] [kubeadm-init] Initializing master0", node.HostInfo.Host)

		output, err := initMaster(node, *c.Kubei, *c.Kubeadm)
		if err != nil {
			return err
		}

		if err := copyAdminConfig(node); err != nil {
			return err
		}

		fmt.Printf("[%s] [kubeadm-init] init master0: %s\n", node.HostInfo.Host, color.HiGreenString("done✅️"))

		klog.V(2).Infof("[%s] [token] Getting token from master init output", node.HostInfo.Host)
		getToken(string(output), &c.Kubernetes.Token)

		return nil
	})
}

func initMaster(node *rundata.Node, kubeiCfg rundata.Kubei, kubeadmCfg rundata.Kubeadm) ([]byte, error) {
	text, err := tmpl.Kubeadm(tmpl.Init, node.Name, kubeiCfg.Kubernetes, kubeadmCfg)
	if err != nil {
		return nil, fmt.Errorf("[%s] [kubeadm-init] Failed to Initialize master0: %v", node.HostInfo.Host, err)
	}

	output, err := node.RunOut(text)
	if err != nil {
		return nil, fmt.Errorf("[%s] [kubeadm-init] Failed to Initialize master0: %v", node.HostInfo.Host, err)
	}
	return output, nil
}

// JoinControlPlane join masters to ControlPlane
func JoinControlPlane(c *rundata.Cluster) error {
	return c.RunOnOtherMastersAndPrintLog(func(node *rundata.Node, c *rundata.Cluster) error {
		apiDomainName, _, _ := net.SplitHostPort(c.Kubeadm.ControlPlaneEndpoint)
		if err := system.SetHost(node, c.ClusterNodes.Masters[0].HostInfo.Host, apiDomainName); err != nil {
			return err
		}

		if err := system.SwapOff(node); err != nil {
			return err
		}

		if err := iptables(node); err != nil {
			return err
		}

		klog.V(3).Infof("[%s] [kubeadm-join] Joining to masters", node.HostInfo.Host)
		if err := joinControlPlane(node, *c.Kubei, *c.Kubeadm); err != nil {
			return err
		}

		fmt.Printf("[%s] [kubeadm-join] join to masters: %s\n", node.HostInfo.Host, color.HiGreenString("done✅️"))

		if err := copyAdminConfig(node); err != nil {
			return err
		}

		return system.SetHost(node, constants.LoopbackAddress, apiDomainName)
	}, color.HiBlueString("Joining to masters ☸️"))
}

func joinControlPlane(node *rundata.Node, kubeiCfg rundata.Kubei, kubeadmCfg rundata.Kubeadm) error {
	text, err := tmpl.Kubeadm(tmpl.JoinControlPlane, node.Name, kubeiCfg.Kubernetes, kubeadmCfg)
	if err != nil {
		return fmt.Errorf("[%s] [kubeadm-join] Failed to join master nodes: %v", node.HostInfo.Host, err)
	}

	if err := node.Run(text); err != nil {
		return fmt.Errorf("[%s] [kubeadm-join] Failed to join master nodes: %v", node.HostInfo.Host, err)
	}

	return nil
}

//getToken get token from kubeadm init output
func getToken(str string, token *rundata.Token) {
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
	if err := node.Run(tmpl.CopyAdminConfig()); err != nil {
		return fmt.Errorf("[%s] [kubectl-config] Failed to copy admin.conf to $HOME/.kube/config: %v", node.HostInfo.Host, err)
	}

	if node.HostInfo.User != "root" {
		klog.V(2).Infof("[%s] [kubectl-config] Chown $HOME/.kube/config to user %s", node.HostInfo.Host, node.HostInfo.User)
		if err := node.Run(tmpl.ChownKubectlConfig()); err != nil {
			return fmt.Errorf("[%s] [kubectl-config] Failed to chown $HOME/.kube/config to user %s: %v", node.HostInfo.Host, node.HostInfo.User, err)
		}
	}

	return nil
}
