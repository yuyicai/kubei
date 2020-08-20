package network

import (
	"fmt"

	"github.com/fatih/color"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/internal/rundata"
)

func Network(c *rundata.Cluster) error {
	switch c.NetworkPlugins.Type {
	case "none":
		color.HiBlue("Does not install network plugin 🌐")
		color.HiYellow("You should install network plugin by yourself after init the kubernetes cluster")
	case "flannel":
		color.HiBlue("Installing flannel network plugin 🌐")
		return Flannel(c)
	case "calico":
		//TODO
		klog.Info("[network] calico //TODO")
	default:
		return fmt.Errorf("[network] Unsupported network type: %s, supported type: calico, flannel, none", c.NetworkPlugins.Type)
	}

	return nil
}
