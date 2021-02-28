package preflight

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/internal/constants"
	"github.com/yuyicai/kubei/internal/operator"
	"github.com/yuyicai/kubei/internal/rundata"
	"github.com/yuyicai/kubei/pkg/ssh"
)

func InitPrepare(c *rundata.Cluster) error {
	color.HiBlue("Checking SSH connect 🌐")
	return operator.RunOnAllNodes(c, func(node *rundata.Node, c *rundata.Cluster) error {
		if err := setSSH(node, c.Kubei); err != nil {
			return err
		}
		if err := checkCommandConntrack(node); err != nil {
			return err
		}
		return checkPackageManagementType(node)
	})
}

func ResetPrepare(c *rundata.Cluster) error {
	color.HiBlue("Checking SSH connect 🌐")
	return operator.RunOnAllNodes(c, func(node *rundata.Node, c *rundata.Cluster) error {
		if err := setSSH(node, c.Kubei); err != nil {
			return err
		}
		return checkPackageManagementType(node)
	})
}

func ExecPrepare(c *rundata.Cluster) error {
	color.HiBlue("Checking SSH connect 🌐")
	return operator.RunOnAllNodes(c, func(node *rundata.Node, c *rundata.Cluster) error {
		return setSSH(node, c.Kubei)
	})
}

func CloseSSH(c *rundata.Cluster) error {
	return operator.RunOnAllNodes(c, func(node *rundata.Node, c *rundata.Cluster) error {
		klog.V(1).Infof("[%s][close] Close ssh connect", node.HostInfo.Host)
		return node.SSH.Close()
	})
}

func setSSH(node *rundata.Node, cfg *rundata.Kubei) error {
	if err := setJumpServer(&cfg.JumpServer); err != nil {
		return fmt.Errorf("[preflight] Failed to set jump server: %v", err)
	}
	if err := setSSHConnect(node, &cfg.JumpServer); err != nil {
		return fmt.Errorf("[%s] [preflight] Failed to set ssh connect: %v", node.HostInfo.Host, err)
	}
	return nil
}

func setJumpServer(jumpServer *rundata.JumpServer) error {
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

func checkCommandConntrack(node *rundata.Node) error {
	// https://github.com/kubernetes/kubernetes/blob/v1.18.0/cmd/kubeadm/app/preflight/checks.go#L1020
	tmp := `
if command -v %s >/dev/null 2>&1; then 
  echo 'true' 
else 
  echo 'false' 
fi
`
	out, err := node.RunOut(fmt.Sprintf(tmp, "conntrack"))
	if err != nil {
		return err
	}

	status := string(out)

	if strings.Contains(status, "true") {
		return nil
	}
	if strings.Contains(status, "false") {
		klog.V(8).Info("conntrack no exists")
		return errors.Errorf("can not find command: %s. you can install conntrack wiht \"yum install -y conntrack (CentOS) or apt update && apt install -y conntrack (Ubuntu)\"", "conntrack")
	}
	klog.V(8).Info("conntrack exists")
	return nil
}

func setSSHConnect(node *rundata.Node, jumpServer *rundata.JumpServer) error {
	if node.SSH == nil {
		return setNodeSSHConnect(node, jumpServer)
	}
	return nil
}

func setNodeSSHConnect(node *rundata.Node, jumpServer *rundata.JumpServer) error {
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

func checkPackageManagementType(node *rundata.Node) error {
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
