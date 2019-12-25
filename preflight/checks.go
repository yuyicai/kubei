package preflight

import (
	"fmt"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/pkg/ssh"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog"
	"strings"
)

func CheckSSH(nodes []*rundata.Node, jumpServer *rundata.JumpServer) error {
	if err := jumpServerCheck(jumpServer); err != nil {
		return fmt.Errorf("[preflight] Failed to set jump server: %v", err)
	}

	g := errgroup.Group{}
	for _, node := range nodes {
		node := node

		g.Go(func() error {
			if err := checkSSH(node, jumpServer); err != nil {
				return fmt.Errorf("[%s] [preflight] Failed to set ssh connect: %v", node.HostInfo.Host, err)
			}
			return nil
		})

	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func checkSSH(node *rundata.Node, jumpServer *rundata.JumpServer) error {
	var err error
	if node.SSH == nil {
		userInfo := node.HostInfo

		//Set up ssh connection through jump server
		if jumpServer.IsUse && jumpServer.Client == nil {
			return fmt.Errorf("Don't set jump server")
		}

		if jumpServer.IsUse {
			klog.Infof("[%s] [preflight] Checking SSH connection (through jump server %s)", userInfo.Host, jumpServer.HostInfo.Host)
			node.SSH, err = ssh.ConnectByJumpServer(userInfo.Host, userInfo.Port, userInfo.User, userInfo.Password, jumpServer.Client)
			if err != nil {
				return err
			}
		} else {

			//Set up ssh connection direct
			klog.Infof("[%s] [preflight] Checking SSH connection", userInfo.Host)
			node.SSH, err = ssh.Connect(userInfo.Host, userInfo.Port, userInfo.User, userInfo.Password)
			if err != nil {
				return err
			}
		}

		klog.V(2).Infof("[%s] [preflight] Checking package management", userInfo.Host)
		output, err := node.SSH.RunOut("cat /proc/version")
		if err != nil {
			return err
		}
		switch true {
		case strings.Contains(string(output), "Ubuntu"):
			klog.V(5).Infof("[%s] [preflight] The package management is \"apt\"", userInfo.Host)
			node.InstallationType = rundata.Apt
		case strings.Contains(string(output), "Red"):
			klog.V(5).Infof("[%s] [preflight] The package management is \"yum\"", userInfo.Host)
			node.InstallationType = rundata.Yum
		default:
			return fmt.Errorf("[%s] [preflight] Unsupported this system", userInfo.Host)
		}
	}
	return nil
}

func jumpServerCheck(jumpServer *rundata.JumpServer) error {
	if jumpServer.IsUse && jumpServer.Client == nil {
		hostInfo := jumpServer.HostInfo
		klog.Infof("[preflight] Checking jump server %s", hostInfo.Host)
		var err error
		jumpServer.Client, err = ssh.Connect(hostInfo.Host, hostInfo.Port, hostInfo.User, hostInfo.Password)
		if err != nil {
			return err
		}
	}
	return nil
}
