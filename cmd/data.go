package cmd

import (
	"github.com/yuyicai/kubei/cmd/phases"
	"github.com/yuyicai/kubei/config/options"
	"github.com/yuyicai/kubei/config/rundata"
)

// runOptions defines all the init options exposed via flags by kubei.
type runOptions struct {
	kubei   *options.Kubei
	kubeadm *options.Kubeadm
}

// compile-time assert that the local data object satisfies the phases data interface.
var _ phases.RunData = &runData{}

// runData defines all the runtime information used when running the kubei workflow;
// this data is shared across all the phases that are included in the workflow.
type runData struct {
	cluster *rundata.Cluster
}

func (d *runData) KubeiCfg() *rundata.Kubei {
	return d.cluster.Kubei
}

func (d *runData) KubeadmCfg() *rundata.Kubeadm {
	return d.cluster.Kubeadm
}

func (d *runData) Cluster() *rundata.Cluster {
	return d.cluster
}
