package kubeadm

import (
	"context"
	"fmt"
	"github.com/bilibili/kratos/pkg/sync/errgroup"
	"github.com/yuyicai/kubei/cmd/tmpl"
	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"k8s.io/klog"
	"net"
	"strings"
)

// InitMaster init master0
func InitMaster(node *rundata.Node, kubeiCfg *rundata.Kubei, kubeadmCfg *rundata.Kubeadm) error {
	apiDomainName, _, _ := net.SplitHostPort(kubeadmCfg.ControlPlaneEndpoint)
	if err := system.SetHost(node, constants.LoopbackAddress, apiDomainName); err != nil {
		return err
	}

	if err := system.SwapOff(node); err != nil {
		return err
	}

	if err := iptables(node); err != nil {
		return err
	}

	klog.Infof("[%s] [kubeadm-init] Initializing master0", node.HostInfo.Host)

	output, err := initMaster(node, *kubeiCfg, *kubeadmCfg)
	if err != nil {
		return err
	}

	if err := copyAdminConfig(node); err != nil {
		return err
	}

	klog.Infof("[%s] [kubeadm-init] Successfully initialized master0", node.HostInfo.Host)

	klog.V(2).Infof("[%s] [token] Getting token from master init output", node.HostInfo.Host)
	getToken(string(output), &kubeiCfg.Kubernetes.Token)

	return nil

}

func initMaster(node *rundata.Node, kubeiCfg rundata.Kubei, kubeadmCfg rundata.Kubeadm) ([]byte, error) {
	text, err := tmpl.Kubeadm(tmpl.Init, node.Name, kubeiCfg.Kubernetes, kubeadmCfg)
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
func JoinControlPlane(masters []*rundata.Node, kubeiCfg *rundata.Kubei, kubeadmCfg *rundata.Kubeadm) error {
	apiDomainName, _, _ := net.SplitHostPort(kubeadmCfg.ControlPlaneEndpoint)
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range masters[1:] {
		node := node
		g.Go(func(ctx context.Context) error {

			if err := system.SetHost(node, masters[0].HostInfo.Host, apiDomainName); err != nil {
				return err
			}

			if err := system.SwapOff(node); err != nil {
				return err
			}

			if err := iptables(node); err != nil {
				return err
			}

			klog.Infof("[%s] [kubeadm-join] Joining master nodes", node.HostInfo.Host)
			if err := joinControlPlane(node, *kubeiCfg, *kubeadmCfg); err != nil {
				return err
			}
			klog.Infof("[%s] [kubeadm-join] Successfully joined master nodes", node.HostInfo.Host)

			if err := copyAdminConfig(node); err != nil {
				return err
			}

			return system.SetHost(node, constants.LoopbackAddress, apiDomainName)
		})
	}

	return g.Wait()

}

func joinControlPlane(node *rundata.Node, kubeiCfg rundata.Kubei, kubeadmCfg rundata.Kubeadm) error {
	text, err := tmpl.Kubeadm(tmpl.JoinControlPlane, node.Name, kubeiCfg.Kubernetes, kubeadmCfg)
	if err != nil {
		return fmt.Errorf("[%s] [kubeadm-join] Failed to join master nodes: %v", node.HostInfo.Host, err)
	}

	if err := node.SSH.Run(text); err != nil {
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
	if err := node.SSH.Run(tmpl.CopyAdminConfig()); err != nil {
		return fmt.Errorf("[%s] [kubectl-config] Failed to copy admin.conf to $HOME/.kube/config: %v", node.HostInfo.Host, err)
	}

	if node.HostInfo.User != "root" {
		klog.V(2).Infof("[%s] [kubectl-config] Chown $HOME/.kube/config to user %s", node.HostInfo.Host, node.HostInfo.User)
		if err := node.SSH.Run(tmpl.ChownKubectlConfig()); err != nil {
			return fmt.Errorf("[%s] [kubectl-config] Failed to chown $HOME/.kube/config to user %s: %v", node.HostInfo.Host, node.HostInfo.User, err)
		}
	}

	return nil
}
