package phases

import (
	"github.com/yuyicai/kubei/config/rundata"
)

type InitData interface {
	Cluster() *rundata.ClusterNodes
	ContainerRuntime() *rundata.ContainerRuntime
	Kube() *rundata.KubeComponent
	Cfg() *rundata.Kubei
	KubeadmCfg() *rundata.Kubeadm
}
