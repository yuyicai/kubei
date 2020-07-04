package kube

import (
	"fmt"

	"k8s.io/klog"

	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"github.com/yuyicai/kubei/tmpl"
)

func InstallKubeComponent(c *rundata.Cluster) error {

	return c.RunOnAllNodes(func(node *rundata.Node) error {
		klog.Infof("[%s] [kube] Installing Kubernetes component", node.HostInfo.Host)
		if err := installKubeComponent(c.Kubernetes.Version, node); err != nil {
			return fmt.Errorf("[%s] [kube] Failed to install Kubernetes component: %v", node.HostInfo.Host, err)
		}

		if err := system.Restart("kubelet", node); err != nil {
			return err
		}
		klog.Infof("[%s] [kube] Successfully installed Kubernetes component", node.HostInfo.Host)

		return nil
	})
}

func installKubeComponent(version string, node *rundata.Node) error {

	cmdTmpl := tmpl.NewKubeText(node.PackageManagementType)
	cmd, err := cmdTmpl.KubeComponent(version, node.InstallType)
	if err != nil {
		return err
	}

	return node.Run(cmd)

}
