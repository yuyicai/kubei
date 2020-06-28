package kube

import (
	"context"
	"fmt"
	"github.com/bilibili/kratos/pkg/sync/errgroup"
	"github.com/yuyicai/kubei/cmd/tmpl"
	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/phases/system"
	"k8s.io/klog"
)

func InstallKubeComponent(version string, nodes []*rundata.Node) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range nodes {
		node := node
		g.Go(func(ctx context.Context) error {
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

	return g.Wait()
}

func installKubeComponent(version string, node *rundata.Node) error {

	cmdTmpl := tmpl.NewKubeText(node.PackageManagementType)
	cmd, err := cmdTmpl.KubeComponent(version, node.InstallType)
	if err != nil {
		return err
	}

	return node.SSH.Run(cmd)

}
