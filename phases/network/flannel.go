package network

import (
	"fmt"

	"k8s.io/klog"

	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/tmpl"
)

func Flannel(c *rundata.Cluster) error {
	return c.RunOnFirstMaster(func(node *rundata.Node) error {
		klog.Infof("[%s] [network] Add the flannel network plugin", node.HostInfo.Host)

		text, err := tmpl.Flannel(c.Kubeadm.Networking.PodSubnet, c.NetworkPlugins.Flannel.Image.GetImage(), c.NetworkPlugins.Flannel.BackendType)
		if err != nil {
			return fmt.Errorf("[%s] [network] Failed to add the flannel network plugin: %v", node.HostInfo.Host, err)
		}

		if err := node.Run(text); err != nil {
			return fmt.Errorf("[%s] [network] Failed to add the flannel network plugin: %v", node.HostInfo.Host, err)
		}
		return nil
	})
}
