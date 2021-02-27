package kube

import (
	"fmt"

	"github.com/fatih/color"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/internal/operator"
	"github.com/yuyicai/kubei/internal/phases/system"
	"github.com/yuyicai/kubei/internal/rundata"
	"github.com/yuyicai/kubei/internal/tmpl"
)

func InstallKubeComponent(c *rundata.Cluster) error {
	color.HiBlue("Installing Kubernetes component ☸️")
	return operator.RunOnAllNodes(c, func(node *rundata.Node, c *rundata.Cluster) error {
		klog.V(2).Infof("[%s] [kube] Installing Kubernetes component", node.HostInfo.Host)
		if err := installKubeComponent(c.Kubernetes.Version, node); err != nil {
			return fmt.Errorf("[%s] [kube] Failed to install Kubernetes component: %v", node.HostInfo.Host, err)
		}

		if err := system.Restart("kubelet", node); err != nil {
			return err
		}
		fmt.Printf("[%s] [kube] install Kubernetes component: %s\n", node.HostInfo.Host, color.HiGreenString("done✅️"))
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
