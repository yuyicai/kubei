package phases

import (
	"github.com/yuyicai/kubei/config/rundata"
)

type ResetData interface {
	Cluster() *rundata.ClusterNodes
	Cfg() *rundata.Kubei
	KubeadmCfg() *rundata.Kubeadm
}
