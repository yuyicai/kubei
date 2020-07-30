package preflight

import (
	"fmt"
	"net"

	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/tmpl"
	"k8s.io/klog"
)

func ResetKubeadm(c *rundata.Cluster) error {
	apiDomainName, _, _ := net.SplitHostPort(c.Kubeadm.ControlPlaneEndpoint)
	return c.RunOnAllNodes(func(node *rundata.Node) error {
		klog.V(2).Infof("[%s] [reset] Resetting node", node.HostInfo.Host)
		if err := resetkubeadmOnNode(node, apiDomainName); err != nil {
			return fmt.Errorf("[%s] [reset] Failed to reset node: %v", node.HostInfo.Host, err)
		}
		klog.Infof("[%s] [reset] Successfully reset node", node.HostInfo.Host)
		return nil
	})
}

func resetkubeadmOnNode(node *rundata.Node, apiDomainName string) error {
	if err := node.Run("yes | kubeadm reset"); err != nil {
		return err
	}

	return node.Run(tmpl.ResetHosts(apiDomainName))
}

func RemoveKubeComponente(c *rundata.Cluster) error {
	return c.RunOnAllNodes(func(node *rundata.Node) error {
		return removeKubeComponente(node)
	})
}

func removeKubeComponente(node *rundata.Node) error {
	klog.V(2).Infof("[%s] [remove] remove the kubernetes component from the node", node.HostInfo.Host)
	if err := removeKubeComponentOnNode(node); err != nil {
		return fmt.Errorf("[%s] [remove] Failed to remove the kubernetes component: %v", node.HostInfo.Host, err)
	}
	klog.Infof("[%s] [remove] Successfully remove the kubernetes component from the node", node.HostInfo.Host)
	return nil
}

func removeKubeComponentOnNode(node *rundata.Node) error {
	cmdTmpl := tmpl.NewKubeText(node.PackageManagementType)
	return node.Run(cmdTmpl.RemoveKubeComponent())
}

func RemoveContainerEngine(c *rundata.Cluster) error {
	return c.RunOnAllNodes(func(node *rundata.Node) error {
		return removeContainerEngine(node)
	})
}

func removeContainerEngine(node *rundata.Node) error {
	klog.V(2).Infof("[%s] [remove] Remove container engine from the node", node.HostInfo.Host)
	if err := removeContainerEngineOnNode(node); err != nil {
		return fmt.Errorf("[%s] [remove] Failed to remove container engine: %v", node.HostInfo.Host, err)
	}
	klog.Infof("[%s] [remove] Successfully remove container engine", node.HostInfo.Host)
	return nil
}

func removeContainerEngineOnNode(node *rundata.Node) error {
	cmdTmpl := tmpl.NewContainerEngineText(node.PackageManagementType)
	return node.Run(cmdTmpl.RemoveDocker())
}
