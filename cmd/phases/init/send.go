package init

import (
	"errors"

	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"github.com/yuyicai/kubei/cmd/phases"
	"github.com/yuyicai/kubei/config/options"
	sendphases "github.com/yuyicai/kubei/phases/send"
)

// NewSendPhase creates a kubei workflow phase that implements handling of send.
func NewSendPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "send",
		Short:        "send kubernetes offline pkg to nodes",
		Long:         "send kubernetes offline pkg to nodes",
		InheritFlags: getSendPhaseFlags(),
		Run:          runSend,
	}
	return phase
}

func getSendPhaseFlags() []string {
	flags := []string{
		options.OfflineFile,
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

func runSend(c workflow.RunData) error {
	data, ok := c.(phases.RunData)
	if !ok {
		return errors.New("runtime phase invoked with an invalid data struct")
	}

	cluster := data.Cluster()

	return sendphases.Send(cluster)
}
