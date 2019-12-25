package phases

import (
	"errors"
	"github.com/yuyicai/kubei/config/options"
	criphases "github.com/yuyicai/kubei/phases/cri"
	"github.com/yuyicai/kubei/preflight"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
)

// NewCriPhase creates a kubei workflow phase that implements handling of cri.
func NewCriPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "cri",
		Short:        "install cri in all nodes",
		Long:         "install cri in all nodes",
		InheritFlags: getCriPhaseFlags(),
		Run:          runCri,
	}
	return phase
}

func getCriPhaseFlags() []string {
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

func runCri(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.New("cri phase invoked with an invalid data struct")
	}

	cfg := data.Cfg()
	nodes := append(cfg.ClusterNodes.Masters, cfg.ClusterNodes.Worker...)

	if err := preflight.CheckSSH(nodes, &cfg.JumpServer); err != nil {
		return err
	}

	if err := criphases.InstallDocker(nodes); err != nil {
		return err
	}

	return nil
}
