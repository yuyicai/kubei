package cri

import (
	"fmt"
	"github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/rundata"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog"
)

func InstallDocker(nodes []*rundata.Node) error {
	g := errgroup.Group{}
	for _, node := range nodes {
		node := node
		g.Go(func() error {
			klog.Infof("[%s] [cri] Installing Docker", node.HostInfo.Host)
			if err := installDocker(node); err != nil {
				return fmt.Errorf("[%s] [cri] Failed to install Docker: %v", node.HostInfo.Host, err)
			}
			klog.Infof("[%s] [cri] Successfully installed Docker", node.HostInfo.Host)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func installDocker(node *rundata.Node) error {
	cmdText := text.NewCriText(node.InstallationType)
	if err := node.SSH.Run(cmdText.Docker()); err != nil {
		return err
	}
	return nil
}
