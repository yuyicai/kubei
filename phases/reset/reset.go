package preflight

import (
	"fmt"
	cmdtext "github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/rundata"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog"
)

func Reset(nodes []*rundata.Node, apiDomainName string) error {

	g := errgroup.Group{}
	for _, node := range nodes {
		node := node
		g.Go(func() error {
			klog.V(2).Infof("[%s] [Reset] Resetting node", node.HostInfo.Host)
			if err := reset(node, apiDomainName); err != nil {
				return fmt.Errorf("[%s] Failed to reset node: %v", node.HostInfo.Host, err)
			}
			klog.Infof("[%s] [Reset] Successfully Reset node", node.HostInfo.Host)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func reset(node *rundata.Node, apiDomainName string) error {
	if err := node.SSH.Run("yes | kubeadm reset"); err != nil {
		return err
	}

	if err := node.SSH.Run(cmdtext.ResetHosts(apiDomainName)); err != nil {
		return err
	}
	return nil
}
