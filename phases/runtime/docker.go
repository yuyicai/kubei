package runtime

import (
	"fmt"
	cmdtext "github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog"
)

func InstallDocker(nodes []*rundata.Node) error {
	g := errgroup.Group{}
	for _, node := range nodes {
		node := node
		g.Go(func() error {
			klog.Infof("[%s] [container-engine] Installing Docker", node.HostInfo.Host)
			if err := installDocker(node); err != nil {
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

func installDocker(node *rundata.Node) error {
	cmdText := cmdtext.NewContainerEngineText(node.InstallationType)
	if err := node.SSH.Run(cmdText.Docker()); err != nil {
		return err
	}
	return nil
}
