package phases

import (
	"errors"
	"github.com/yuyicai/kubei/config/options"
	kubephases "github.com/yuyicai/kubei/phases/kube"
	"github.com/yuyicai/kubei/preflight"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
)

// NewKubeComponentPhase creates a kubei workflow phase that implements handling of kube.
func NewKubeComponentPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "kube",
		Short:        "install kube in all nodes",
		Long:         "install kube in all nodes",
		InheritFlags: getKubeComponentPhaseFlags(),
		Run:          runKubeComponent,
	}
	return phase
}

func getKubeComponentPhaseFlags() []string {
	flags := []string{
		options.JumpServer,
		options.KubernetesVersion,
		options.Masters,
		options.Workers,
		options.Password,
		options.Port,
		options.User,
	}
	return flags
}

func runKubeComponent(c workflow.RunData) error {
	data, ok := c.(InitData)
	if !ok {
		return errors.New("kube phase invoked with an invalid data struct")
	}

	cfg := data.Cfg()
	nodes := append(cfg.ClusterNodes.Masters, cfg.ClusterNodes.Worker...)

	if err := preflight.CheckSSH(nodes, &cfg.JumpServer); err != nil {
		return err
	}

	if err := kubephases.InstallKubeComponent(nodes); err != nil {
		return err
	}

	return nil
}
