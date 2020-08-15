package network

import (
	"fmt"

	"k8s.io/klog"

	"github.com/yuyicai/kubei/internal/rundata"
)

func Network(c *rundata.Cluster) error {
	switch c.NetworkPlugins.Type {
	case "none":
		klog.Info("[network] Does not network plugin")
	case "flannel":
		return Flannel(c)
	case "calico":
		//TODO
		klog.Info("[network] calico //TODO")
	default:
		return fmt.Errorf("[network] Unsupported network type: %s, supported type: calico, flannel, none", c.NetworkPlugins.Type)
	}

	return nil
}
