package reset

import (
	"errors"
	"github.com/yuyicai/kubei/cmd/phases"
	"github.com/yuyicai/kubei/config/options"
	resetphases "github.com/yuyicai/kubei/phases/reset"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
)

// NewResetPhase creates a kubei workflow phase that implements handling of kubernetes-component.
func NewKubeComponentPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "kubernetes-component",
		Short:        "remove the kubernetes component from the nodes",
		Long:         "remove the kubernetes component from the nodes",
		InheritFlags: getKubeComponentPhaseFlags(),
		Run:          runKubeComponent,
	}
	return phase
}

func getKubeComponentPhaseFlags() []string {
	flags := []string{
		options.RemoveKubernetesComponent,
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

func runKubeComponent(c workflow.RunData) error {
	data, ok := c.(phases.RunData)
	if !ok {
		return errors.New("reset phase invoked with an invalid rundata struct")
	}

	cfg := data.KubeiCfg()

	cluster := data.Cluster()

	if cfg.Reset.RemoveKubeComponent {
		return resetphases.RemoveKubeComponente(cluster)
	}

	return nil
}
