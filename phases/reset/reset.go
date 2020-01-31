package preflight

import (
	"context"
	"fmt"
	"github.com/bilibili/kratos/pkg/sync/errgroup"
	cmdtext "github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/rundata"
	"k8s.io/klog"
)

func ResetKubeadm(nodes []*rundata.Node, apiDomainName string) error {

	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(20)
	for _, node := range nodes {
		node := node
		g.Go(func(ctx context.Context) error {
			klog.V(2).Infof("[%s] [reset] Resetting node", node.HostInfo.Host)
			if err := resetKubeadm(node, apiDomainName); err != nil {
				return fmt.Errorf("[%s] [reset] Failed to reset node: %v", node.HostInfo.Host, err)
			}
			klog.Infof("[%s] [reset] Successfully reset node", node.HostInfo.Host)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func resetKubeadm(node *rundata.Node, apiDomainName string) error {
	if err := node.SSH.Run("yes | kubeadm reset"); err != nil {
		return err
	}

	if err := node.SSH.Run(cmdtext.ResetHosts(apiDomainName)); err != nil {
		return err
	}
	return nil
}

func RemoveKubeComponente(nodes []*rundata.Node) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(20)
	for _, node := range nodes {
		node := node
		g.Go(func(ctx context.Context) error {
			klog.V(2).Infof("[%s] [remove] remove the kubernetes component from the node", node.HostInfo.Host)
			if err := removeKubeComponent(node); err != nil {
				return fmt.Errorf("[%s] [remove] Failed to remove the kubernetes component: %v", node.HostInfo.Host, err)
			}
			klog.Infof("[%s] [remove] Successfully remove the kubernetes component from the node", node.HostInfo.Host)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func removeKubeComponent(node *rundata.Node) error {
	cmdText := cmdtext.NewKubeText(node.InstallationType)
	if err := node.SSH.Run(cmdText.RemoveKubeComponent()); err != nil {
		return err
	}
	return nil
}

func RemoveContainerEngine(nodes []*rundata.Node) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(20)
	for _, node := range nodes {
		node := node
		g.Go(func(ctx context.Context) error {
			klog.V(2).Infof("[%s] [remove] Remove container engine from the node", node.HostInfo.Host)
			if err := removeContainerEngine(node); err != nil {
				return fmt.Errorf("[%s] [remove] Failed to remove container engine: %v", node.HostInfo.Host, err)
			}
			klog.Infof("[%s] [remove] Successfully remove container engine", node.HostInfo.Host)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func removeContainerEngine(node *rundata.Node) error {
	cmdText := cmdtext.NewContainerEngineText(node.InstallationType)
	if err := node.SSH.Run(cmdText.RemoveDocker()); err != nil {
		return err
	}
	return nil
}
