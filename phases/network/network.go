package network

import (
	"fmt"
	"github.com/yuyicai/kubei/config/rundata"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

func Network(node *rundata.Node, net rundata.Network, knet kubeadmapi.Networking) error {
	switch net.Type {
	case "none":
		klog.Info("[network] Does not network plugin")
	case "flannel":
		if err := Flannel(node, net.Flannel, knet); err != nil {
			return err
		}
	case "calico":
		//TODO
		klog.Info("[network] calico //TODO")
	default:
		return fmt.Errorf("[network] Unsupported network type: %s, supported type: calico, flannel, none", net.Type)
	}

	return nil
}
