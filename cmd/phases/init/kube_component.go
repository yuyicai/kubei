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
		Short:        "install Kubernetes component",
		Long:         "install Kubernetes component",
		InheritFlags: getKubeComponentPhaseFlags(),
		Run:          runKubeComponent,
	}
	return phase
}

func getKubeComponentPhaseFlags() []string {
	flags := []string{
		options.OfflineFile,
		options.JumpServer,
		options.KubernetesVersion,
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
	data, ok := c.(InitData)
	if !ok {
		return errors.New("kube phase invoked with an invalid data struct")
	}

	cfg := data.KubeiCfg()

	if err := preflight.Check(cfg); err != nil {
		return err
	}

	return kubephases.InstallKubeComponent(cfg.Kubernetes.Version, cfg.ClusterNodes.GetAllNodes())

}
