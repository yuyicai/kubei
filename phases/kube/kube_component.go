package kube

import (
	"fmt"
	cmdtext "github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog"
)

func InstallKubeComponent(version string, nodes []*rundata.Node) error {
	g := errgroup.Group{}
	for _, node := range nodes {
		node := node
		g.Go(func() error {
			klog.Infof("[%s] [kube] Installing Kubernetes component", node.HostInfo.Host)
			if err := installKubeComponent(version, node); err != nil {
				return fmt.Errorf("[%s] [kube] Failed to install Kubernetes component: %v", node.HostInfo.Host, err)
			}

			if err := system.Restart("kubelet", node); err != nil {
				return err
			}

			klog.Infof("[%s] [kube] Successfully installed Kubernetes component", node.HostInfo.Host)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func installKubeComponent(version string, node *rundata.Node) error {
	cmdText := cmdtext.NewKubeText(node.InstallationType)
	cmd, err := cmdText.KubeComponent(version)
	if err != nil {
		return err
	}
	if err := node.SSH.Run(cmd); err != nil {
		return err
	}
	return nil
}
