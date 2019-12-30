package runtime

import (
	"fmt"
	cmdtext "github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog"
)

func InstallDocker(version string, nodes []*rundata.Node) error {
	g := errgroup.Group{}
	for _, node := range nodes {
		node := node
		g.Go(func() error {
			klog.Infof("[%s] [container-engine] Installing Docker", node.HostInfo.Host)
			if err := installDocker(version, node); err != nil {
				return fmt.Errorf("[%s] [container-engine] Failed to install Docker: %v", node.HostInfo.Host, err)
			}

			if err := system.Restart("docker", node); err != nil {
				return err
			}

			klog.Infof("[%s] [container-engine] Successfully installed Docker", node.HostInfo.Host)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func installDocker(version string, node *rundata.Node) error {
	cmdText := cmdtext.NewContainerEngineText(node.InstallationType)
	cmd, err := cmdText.Docker(version)
	if err != nil {
		return err
	}
	if err := node.SSH.Run(cmd); err != nil {
		return err
	}
	return nil
}
