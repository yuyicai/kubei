package phases

import (
	"errors"
	"github.com/yuyicai/kubei/config/options"
	resetphases "github.com/yuyicai/kubei/phases/reset"
	"github.com/yuyicai/kubei/preflight"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
)

// NewResetPhase creates a kubei workflow phase that implements handling of kubeadm.
func NewContainerEnginePhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "container-engine",
		Short:        "remove the container engine from the nodes",
		Long:         "remove the container engine from the nodes",
		InheritFlags: getContainerEnginePhaseFlags(),
		Run:          runContainerEngine,
	}
	return phase
}

func getContainerEnginePhaseFlags() []string {
	flags := []string{
		options.RemoveContainerEngine,
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

func runContainerEngine(c workflow.RunData) error {
	data, ok := c.(ResetData)
	if !ok {
		return errors.New("reset phase invoked with an invalid rundata struct")
	}

	cfg := data.Cfg()

	if cfg.Reset.RemoveContainerEngine {
		nodes := append(cfg.ClusterNodes.Masters, cfg.ClusterNodes.Worker...)

		if err := preflight.Check(nodes, &cfg.JumpServer); err != nil {
			return err
		}

		if err := resetphases.RemoveContainerEngine(nodes); err != nil {
			return err
		}
	}

	return nil
}
