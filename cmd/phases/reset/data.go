package phases

import (
	"github.com/yuyicai/kubei/config/rundata"
)

type ResetData interface {
	KubeiCfg() *rundata.Kubei
	KubeadmCfg() *rundata.Kubeadm
}
