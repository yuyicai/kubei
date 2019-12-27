package network

import (
	"fmt"
	cmdtext "github.com/yuyicai/kubei/cmd/text"
	"github.com/yuyicai/kubei/config/rundata"
	"k8s.io/klog"
)

func Flannel(node *rundata.Node, network, image, backendType string) error {
	klog.Infof("[%s] [network] Add the flannel network plugin", node.HostInfo.Host)

	text, err := cmdtext.Flannel(network, image, backendType)
	if err != nil {
		return fmt.Errorf("[%s] [network] Failed to add the flannel network plugin: %v", node.HostInfo.Host, err)
	}

	if err := node.SSH.Run(text); err != nil {
		return fmt.Errorf("[%s] [network] Failed to add the flannel network plugin: %v", node.HostInfo.Host, err)
	}
	return nil
}
