package send

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/fatih/color"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/internal/constants"
	"github.com/yuyicai/kubei/internal/rundata"
)

func Send(c *rundata.Cluster) error {
	color.HiBlue("Sending Kubernetes offline pkg to nodes ✉️")
	return c.RunOnAllNodes(func(node *rundata.Node, c *rundata.Cluster) error {
		if err := send(node, c.Kubei); err != nil {
			return err
		}

		fmt.Printf("[%s] [send] send kubernetes offline pkg: %s\n", node.HostInfo.Host, color.HiGreenString("done✅️"))
		return nil
	})
}

func send(node *rundata.Node, cfg *rundata.Kubei) error {
	return sendAndtar(path.Join("/tmp/.kubei", filepath.Base(cfg.OfflineFile)), cfg.OfflineFile, node)
}

func sendAndtar(dstFile, srcFile string, node *rundata.Node) error {
	if node.InstallType == constants.InstallTypeOffline && !node.IsSend {
		if err := sendFile(dstFile, srcFile, node); err != nil {
			return err
		}
		klog.V(3).Infof("[%s] [send] send pkg to %s, ", node.HostInfo.Host, dstFile)
		if err := tar(dstFile, node); err != nil {
			return fmt.Errorf("[%s] [tar] failed to Decompress the file %s: %v", node.HostInfo.Host, dstFile, err)
		}
		node.IsSend = true
		return nil
	}

	return nil
}

func sendFile(dstFile, srcFile string, node *rundata.Node) error {
	return node.SSH.SendFile(dstFile, srcFile)
}

func tar(file string, node *rundata.Node) error {
	return node.Run(fmt.Sprintf("tar xf %s -C /tmp/.kubei", file))
}
