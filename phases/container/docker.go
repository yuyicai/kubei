package container

import (
	"fmt"

	"k8s.io/klog"

	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"github.com/yuyicai/kubei/tmpl"
)

func InstallDocker(c *rundata.Cluster) error {

	return c.RunOnAllNodes(func(node *rundata.Node) error {
		klog.Infof("[%s] [container-engine] Installing Docker", node.HostInfo.Host)
		if err := installDocker(node, c.ContainerEngine.Docker); err != nil {
			return fmt.Errorf("[%s] [container-engine] Failed to install Docker: %v", node.HostInfo.Host, err)
		}

		if err := system.Restart("docker", node); err != nil {
			return err
		}

		klog.Infof("[%s] [container-engine] Successfully installed Docker", node.HostInfo.Host)
		return nil
	})
}

func installDocker(node *rundata.Node, d rundata.Docker) error {
	cmdTmpl := tmpl.NewContainerEngineText(node.PackageManagementType)
	cmd, err := cmdTmpl.Docker(node.InstallType, d)
	if err != nil {
		return err
	}

	return node.Run(cmd)
}
