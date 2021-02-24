package preflight

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/internal/constants"
	"github.com/yuyicai/kubei/internal/rundata"
	"github.com/yuyicai/kubei/pkg/ssh"
)

func Prepare(c *rundata.Cluster) error {
	color.HiBlue("Checking SSH connect 🌐")
	return c.RunOnAllNodes(func(node *rundata.Node, c *rundata.Cluster) error {
		return check(node, c.Kubei)
	})
}

func CloseSSH(c *rundata.Cluster) error {
	return c.RunOnAllNodes(func(node *rundata.Node, c *rundata.Cluster) error {
		klog.V(1).Infof("[%s][close] Close ssh connect", node.HostInfo.Host)
		return node.SSH.Close()
	})
}

func check(node *rundata.Node, cfg *rundata.Kubei) error {
	if err := jumpServerCheck(&cfg.JumpServer); err != nil {
		return fmt.Errorf("[preflight] Failed to set jump server: %v", err)
	}
	return nodesCheck(node, cfg)
}

func jumpServerCheck(jumpServer *rundata.JumpServer) error {
	if jumpServer.HostInfo.Host != "" && jumpServer.Client == nil {
		hostInfo := jumpServer.HostInfo
		klog.V(5).Infof("[preflight] Checking jump server %s", hostInfo.Host)
		var err error
		jumpServer.Client, err = ssh.Connect(hostInfo.Host, hostInfo.Port, hostInfo.User, hostInfo.Password, hostInfo.Key)
		if err != nil {
			return err
		}
		fmt.Printf("[%s] [preflight] jump server SSH connect: %s\n", hostInfo.Host, color.HiGreenString("done✅️"))
		return nil
	}

	return nil
}

func nodesCheck(node *rundata.Node, cfg *rundata.Kubei) error {

	if err := sshCheck(node, &cfg.JumpServer); err != nil {
		return fmt.Errorf("[%s] [preflight] Failed to set ssh connect: %v", node.HostInfo.Host, err)
	}

	return packageManagementTypeCheck(node)
}

func sshCheck(node *rundata.Node, jumpServer *rundata.JumpServer) error {
	if node.SSH == nil {
		return setSSHConnect(node, jumpServer)
	}
	return nil
}

func setSSHConnect(node *rundata.Node, jumpServer *rundata.JumpServer) error {
	var err error
	userInfo := node.HostInfo
	//Set up ssh connection through jump server
	if jumpServer.HostInfo.Host != "" {
		fmt.Printf("[%s] [preflight] SSH connect (through jump server %s\n): %s", userInfo.Host, jumpServer.HostInfo.Host, color.HiGreenString("done✅️"))
		node.SSH, err = ssh.ConnectByJumpServer(userInfo.Host, userInfo.Port, userInfo.User, userInfo.Password, userInfo.Key, jumpServer.Client)
		return err
	} else {
		//Set up ssh connection direct
		fmt.Printf("[%s] [preflight] SSH connect: %s\n", userInfo.Host, color.HiGreenString("done✅️"))
		node.SSH, err = ssh.Connect(userInfo.Host, userInfo.Port, userInfo.User, userInfo.Password, userInfo.Key)
		return err
	}
}

func packageManagementTypeCheck(node *rundata.Node) error {
	hostInfo := node.HostInfo

	klog.V(2).Infof("[%s] [preflight] Checking package management", hostInfo.Host)
	output, err := node.RunOut("cat /proc/version")
	if err != nil {
		return err
	}

	outputStr := string(output)
	switch true {
	case strings.Contains(outputStr, "Debian"):
		klog.V(5).Infof("[%s] [preflight] The package management is \"apt\"", hostInfo.Host)
		node.PackageManagementType = constants.PackageManagementTypeApt
	case strings.Contains(outputStr, "Ubuntu"):
		klog.V(5).Infof("[%s] [preflight] The package management is \"apt\"", hostInfo.Host)
		node.PackageManagementType = constants.PackageManagementTypeApt
	case strings.Contains(outputStr, "Red"):
		klog.V(5).Infof("[%s] [preflight] The package management is \"yum\"", hostInfo.Host)
		node.PackageManagementType = constants.PackageManagementTypeYum
	default:
		return fmt.Errorf("[%s] [preflight] Unsupported this system", hostInfo.Host)
	}
	return nil
}
