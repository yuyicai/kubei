package phases

import (
	"github.com/yuyicai/kubei/internal/rundata"
)

type RunData interface {
	KubeiCfg() *rundata.Kubei
	KubeadmCfg() *rundata.Kubeadm
	Cluster() *rundata.Cluster
}
