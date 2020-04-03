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

func Check(cfg *rundata.Kubei) error {

	if err := jumpServerCheck(&cfg.JumpServer); err != nil {
		return fmt.Errorf("[preflight] Failed to set jump server: %v", err)
	}

	return nodesCheck(cfg)
}

func jumpServerCheck(jumpServer *rundata.JumpServer) error {
	if jumpServer.HostInfo.Host != "" && jumpServer.Client == nil {
		hostInfo := jumpServer.HostInfo
		klog.Infof("[preflight] Checking jump server %s", hostInfo.Host)
		var err error
		jumpServer.Client, err = ssh.Connect(hostInfo.Host, hostInfo.Port, hostInfo.User, hostInfo.Password, hostInfo.Key)
		return err
	}

	return nil
}

func nodesCheck(cfg *rundata.Kubei) error {

	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range cfg.ClusterNodes.GetAllNodes() {
		node := node
		g.Go(func(ctx context.Context) error {

			if err := sshCheck(node, &cfg.JumpServer); err != nil {
				return fmt.Errorf("[%s] [preflight] Failed to set ssh connect: %v", node.HostInfo.Host, err)
			}

			if err := packageManagementTypeCheck(node); err != nil {
				return err
			}

			return sendAndtar("", "", node)
		})
	}

	return g.Wait()
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
		klog.Infof("[%s] [preflight] Checking SSH connection (through jump server %s)", userInfo.Host, jumpServer.HostInfo.Host)
		node.SSH, err = ssh.ConnectByJumpServer(userInfo.Host, userInfo.Port, userInfo.User, userInfo.Password, userInfo.Key, jumpServer.Client)
		return err
	} else {
		//Set up ssh connection direct
		klog.Infof("[%s] [preflight] Checking SSH connection", userInfo.Host)
		node.SSH, err = ssh.Connect(userInfo.Host, userInfo.Port, userInfo.User, userInfo.Password, userInfo.Key)
		return err
	}
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

func sendAndtar(dstFile, srcFile string, node *rundata.Node) error {
	if node.InstallType == constants.InstallTypeOffline && !node.IsSend {
		if err := sendFile(dstFile, srcFile, node); err != nil {
			return err
		}
		klog.Infof("[%s] [send] send %s, ", node.HostInfo.Host, "dstFile")

		return tar(dstFile, node)
	}
	node.IsSend = true
	return nil
}

func sendFile(dstFile, srcFile string, node *rundata.Node) error {
	return node.SSH.SendFile(dstFile, srcFile)
}

func tar(dstFile string, node *rundata.Node) error {
	return node.SSH.Run(fmt.Sprintf("mkdir -p /tmp/.kube && tar xf %s -C /tmp/.kube", dstFile))
}
