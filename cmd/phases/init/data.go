package phases

import (
	"github.com/yuyicai/kubei/config/rundata"
)

type InitData interface {
	KubeiCfg() *rundata.Kubei
	KubeadmCfg() *rundata.Kubeadm
}
