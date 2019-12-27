package phases

import (
	"github.com/yuyicai/kubei/config/rundata"
)

type InitData interface {
	Cluster() *rundata.ClusterNodes
	ContainerEngine() *rundata.ContainerEngine
	Kube() *rundata.KubeComponent
	Cfg() *rundata.Kubei
	KubeadmCfg() *rundata.Kubeadm
}
