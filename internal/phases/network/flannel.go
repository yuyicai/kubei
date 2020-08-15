package network

import (
	"fmt"

	"github.com/fatih/color"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/internal/rundata"
	"github.com/yuyicai/kubei/internal/tmpl"
)

func Flannel(c *rundata.Cluster) error {
	return c.RunOnFirstMaster(func(node *rundata.Node) error {
		klog.V(3).Infof("[%s] [network] Add the flannel network plugin", node.HostInfo.Host)

		text, err := tmpl.Flannel(c.Kubeadm.Networking.PodSubnet, c.NetworkPlugins.Flannel.Image.GetImage(), c.NetworkPlugins.Flannel.BackendType)
		if err != nil {
			return fmt.Errorf("[%s] [network] Failed to add the flannel network plugin: %v", node.HostInfo.Host, err)
		}

		if err := node.Run(text); err != nil {
			return fmt.Errorf("[%s] [network] Failed to add the flannel network plugin: %v", node.HostInfo.Host, err)
		}

		fmt.Printf("[%s] [network] Add the flannel network plugin: %s\n", node.HostInfo.Host, color.HiGreenString("done✅️"))
		return nil
	})
}
