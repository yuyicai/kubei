package init

import (
	"errors"
	"github.com/yuyicai/kubei/cmd/phases"

	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"github.com/yuyicai/kubei/internal/options"
	containerphases "github.com/yuyicai/kubei/internal/phases/container"
)

// NewContainerEnginePhase creates a kubei workflow phase that implements handling of container engine.
func NewContainerEnginePhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "container-engine",
		Short:        "install container engine",
		Long:         "install container engine",
		InheritFlags: getContainerEnginePhaseFlags(),
		Run:          runContainerEngine,
	}
	return phase
}

func getContainerEnginePhaseFlags() []string {
	flags := []string{
		options.OfflineFile,
		options.JumpServer,
		options.ContainerEngineVersion,
		options.Masters,
		options.Workers,
		options.Password,
		options.Port,
		options.User,
		options.Key,
	}
	return flags
}

func runContainerEngine(c workflow.RunData) error {
	data, ok := c.(phases.RunData)
	if !ok {
		return errors.New("runtime phase invoked with an invalid data struct")
	}

	cluster := data.Cluster()

	return containerphases.InstallContainerEngine(cluster)
}
