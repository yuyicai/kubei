package phases

import (
	"errors"
	"github.com/yuyicai/kubei/config/options"
	runtimephases "github.com/yuyicai/kubei/phases/runtime"
	"github.com/yuyicai/kubei/preflight"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
)

// NewcontainerEnginePhase creates a kubei workflow phase that implements handling of runtime.
func NewcontainerEnginePhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "runtime",
		Short:        "install container runtime",
		Long:         "install container runtime",
		InheritFlags: getcontainerEnginePhaseFlags(),
		Run:          runcontainerEngine,
	}
	return phase
}

func getcontainerEnginePhaseFlags() []string {
	flags := []string{
		options.JumpServer,
		options.DockerVersion,
		options.Masters,
		options.Workers,
		options.Password,
		options.Port,
		options.User,
	}
	return flags
}

func runcontainerEngine(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.New("runtime phase invoked with an invalid data struct")
	}

	cfg := data.Cfg()
	nodes := append(cfg.ClusterNodes.Masters, cfg.ClusterNodes.Worker...)

	if err := preflight.CheckSSH(nodes, &cfg.JumpServer); err != nil {
		return err
	}

	if err := runtimephases.InstallDocker(nodes); err != nil {
		return err
	}

	return nil
}
