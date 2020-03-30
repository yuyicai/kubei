package preflight

import (
	"context"
	"fmt"
	"strings"

	"github.com/bilibili/kratos/pkg/sync/errgroup"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/pkg/ssh"
)

func Check(nodes []*rundata.Node, jumpServer *rundata.JumpServer) error {
	if err := jumpServerCheck(jumpServer); err != nil {
		return fmt.Errorf("[preflight] Failed to set jump server: %v", err)
	}

	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(20)
	for _, node := range nodes {
		node := node
		g.Go(func(ctx context.Context) error {
			if err := sshCheck(node, jumpServer); err != nil {
				return fmt.Errorf("[%s] [preflight] Failed to set ssh connect: %v", node.HostInfo.Host, err)
			}

			if err := packageManagementTypeCheck(node); err != nil {
				return err
			}

			return nil
		})

	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func sshCheck(node *rundata.Node, jumpServer *rundata.JumpServer) error {
	if node.SSH == nil {
		return setSSHConnect(node, jumpServer)
	}
	return nil
}

func jumpServerCheck(jumpServer *rundata.JumpServer) error {
	if jumpServer.HostInfo.Host != "" && jumpServer.Client == nil {
		hostInfo := jumpServer.HostInfo
		klog.Infof("[preflight] Checking jump server %s", hostInfo.Host)
		var err error
		jumpServer.Client, err = ssh.Connect(hostInfo.Host, hostInfo.Port, hostInfo.User, hostInfo.Password, hostInfo.Key)
		if err != nil {
			return err
		}
	}
	return nil
}

func setSSHConnect(node *rundata.Node, jumpServer *rundata.JumpServer) error {
	var err error
	userInfo := node.HostInfo

	//Set up ssh connection through jump server
	if jumpServer.HostInfo.Host != "" {
		klog.Infof("[%s] [preflight] Checking SSH connection (through jump server %s)", userInfo.Host, jumpServer.HostInfo.Host)
		node.SSH, err = ssh.ConnectByJumpServer(userInfo.Host, userInfo.Port, userInfo.User, userInfo.Password, userInfo.Key, jumpServer.Client)
		if err != nil {
			return err
		}
	} else {

		//Set up ssh connection direct
		klog.Infof("[%s] [preflight] Checking SSH connection", userInfo.Host)
		node.SSH, err = ssh.Connect(userInfo.Host, userInfo.Port, userInfo.User, userInfo.Password, userInfo.Key)
		if err != nil {
			return err
		}
	}
	return nil
}

func packageManagementTypeCheck(node *rundata.Node) error {
	hostInfo := node.HostInfo

	klog.V(2).Infof("[%s] [preflight] Checking package management", hostInfo.Host)
	output, err := node.SSH.RunOut("cat /proc/version")
	if err != nil {
		return err
	}

	outputStr := string(output)
	switch true {
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

func offlineCheck(node *rundata.Node, install rundata.Install) error {
	if install.Type == constants.InstallTypeOffline {
		node.IsOffline = true
	}
	return nil
}
