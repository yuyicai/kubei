package phases

import (
	"github.com/yuyicai/kubei/config/rundata"
)

type RunData interface {
	KubeiCfg() *rundata.Kubei
	KubeadmCfg() *rundata.Kubeadm
	Cluster() *rundata.Cluster
}
