package system

import (
	"fmt"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/tmpl"
	"k8s.io/klog"
)

func SetHost(node *rundata.Node, ip, apiDomainName string) error {
	klog.V(2).Infof("[%s] [host] Add \"%s %s\" to /etc/hosts", node.HostInfo.Host, ip, apiDomainName)
	if err := node.Run(tmpl.SetHosts(ip, apiDomainName)); err != nil {
		return fmt.Errorf("[%s] [host] Failed to set /etc/hosts: %v", node.HostInfo.Host, err)
	}
	return nil
}

func SwapOff(node *rundata.Node) error {
	klog.V(2).Infof("[%s] [swap] Disable swap", node.HostInfo.Host)
	if err := node.Run(tmpl.SwapOff()); err != nil {
		return fmt.Errorf("[%s] [swap]  Failed to disable swap: %v", node.HostInfo.Host, err)
	}
	return nil
}

func Restart(name string, node *rundata.Node) error {
	klog.V(2).Infof("[%s] [restart] Restart %s", node.HostInfo.Host, name)
	if err := node.Run(tmpl.Restart(name)); err != nil {
		return fmt.Errorf("[%s] [restart] Failed to restart %s: %v", node.HostInfo.Host, name, err)
	}
	return nil
}
