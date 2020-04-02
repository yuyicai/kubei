package phases

import (
	"errors"
	"github.com/yuyicai/kubei/config/options"
	runtimephases "github.com/yuyicai/kubei/phases/runtime"
	"github.com/yuyicai/kubei/preflight"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
)

// NewContainerEnginePhase creates a kubei workflow phase that implements handling of runtime.
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
	data, ok := c.(InitData)
	if !ok {
		return errors.New("runtime phase invoked with an invalid data struct")
	}

	cfg := data.Cfg()
	containerEngine := data.ContainerEngine()
	nodes := cfg.ClusterNodes.GetAllNodes()

	if err := preflight.Check(nodes, &cfg.JumpServer); err != nil {
		return err
	}

	return runtimephases.InstallContainerEngine(nodes, *containerEngine)
}
