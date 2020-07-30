package send

import (
	"fmt"
	"path"
	"path/filepath"

	"k8s.io/klog"

	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
)

func Send(c *rundata.Cluster) error {
	return c.RunOnAllNodes(func(node *rundata.Node) error {
		return send(node, c.Kubei)
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
		klog.Infof("[%s] [send] send pkg to %s, ", node.HostInfo.Host, dstFile)
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
