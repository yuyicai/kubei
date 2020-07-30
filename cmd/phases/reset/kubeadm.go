package reset

import (
	"errors"
	"github.com/yuyicai/kubei/cmd/phases"
	"github.com/yuyicai/kubei/config/options"
	resetphases "github.com/yuyicai/kubei/phases/reset"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
)

// NewResetPhase creates a kubei workflow phase that implements handling of cluster.
func NewKubeadmPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "cluster",
		Short:        "reset Kubernetes cluster",
		Long:         "reset Kubernetes cluster",
		InheritFlags: getKubeadmPhaseFlags(),
		Run:          runKubeadm,
	}
	return phase
}

func getKubeadmPhaseFlags() []string {
	flags := []string{
		options.JumpServer,
		options.Masters,
		options.Workers,
		options.Password,
		options.Port,
		options.User,
		options.Key,
	}
	return flags
}

func runKubeadm(c workflow.RunData) error {
	data, ok := c.(phases.RunData)
	if !ok {
		return errors.New("reset phase invoked with an invalid rundata struct")
	}

	cluster := data.Cluster()

	return resetphases.ResetKubeadm(cluster)
}
