package init

import (
	"errors"
	"github.com/yuyicai/kubei/cmd/phases"

	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"github.com/yuyicai/kubei/config/options"
	kubephases "github.com/yuyicai/kubei/phases/kube"
	"github.com/yuyicai/kubei/preflight"
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
	data, ok := c.(phases.RunData)
	if !ok {
		return errors.New("kube phase invoked with an invalid data struct")
	}

	cluster := data.Cluster()

	if err := preflight.Prepare(cluster); err != nil {
		return err
	}

	return kubephases.InstallKubeComponent(cluster)

}
