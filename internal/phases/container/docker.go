package container

import (
	"fmt"

	"github.com/fatih/color"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/internal/phases/system"
	"github.com/yuyicai/kubei/internal/rundata"
	"github.com/yuyicai/kubei/internal/tmpl"
)

func InstallDocker(c *rundata.Cluster) error {

	color.HiBlue("Installing Docker on all nodes 🐳")
	return c.RunOnAllNodes(func(node *rundata.Node, c *rundata.Cluster) error {
		klog.V(2).Infof("[%s] [container-engine] Installing Docker", node.HostInfo.Host)
		if err := installDocker(node, c.ContainerEngine.Docker); err != nil {
			return fmt.Errorf("[%s] [container-engine] Failed to install Docker: %v", node.HostInfo.Host, err)
		}

		if err := system.Restart("docker", node); err != nil {
			return err
		}
		fmt.Printf("[%s] [container-engine] install Docker: %s\n", node.HostInfo.Host, color.HiGreenString("done✅️"))
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
