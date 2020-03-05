package network

import (
	"fmt"
	cmdtext "github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/rundata"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

func Flannel(node *rundata.Node, f rundata.Flannel, net kubeadmapi.Networking) error {
	klog.Infof("[%s] [network] Add the flannel network plugin", node.HostInfo.Host)

	text, err := cmdtext.Flannel(net.PodSubnet, f.Image.Flannel, f.BackendType)
	if err != nil {
		return fmt.Errorf("[%s] [network] Failed to add the flannel network plugin: %v", node.HostInfo.Host, err)
	}

	if err := node.SSH.Run(text); err != nil {
		return fmt.Errorf("[%s] [network] Failed to add the flannel network plugin: %v", node.HostInfo.Host, err)
	}
	return nil
}
