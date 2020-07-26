package init

import (
	"errors"

	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"github.com/yuyicai/kubei/cmd/phases"
	"github.com/yuyicai/kubei/config/options"
	certphases "github.com/yuyicai/kubei/phases/cert"
	"github.com/yuyicai/kubei/preflight"
)

// NewCertPhase creates a kubei workflow phase that implements handling of cert.
func NewCertPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "cert",
		Short:        "create k8s cluster cert and kubeconfig",
		Long:         "create k8s cluster cert and kubeconfig",
		InheritFlags: getCertPhaseFlags(),
		Run:          runCert,
	}
	return phase
}

func getCertPhaseFlags() []string {
	flags := []string{
		options.JumpServer,
		options.ControlPlaneEndpoint,
		options.ServiceCidr,
		options.Masters,
		options.Workers,
		options.Password,
		options.Port,
		options.User,
		options.Key,
	}
	return flags
}

func runCert(c workflow.RunData) error {
	data, ok := c.(phases.RunData)
	if !ok {
		return errors.New("cert phase invoked with an invalid rundata struct")
	}

	cluster := data.Cluster()

	if err := preflight.Prepare(cluster); err != nil {
		return err
	}

	if err := certphases.CreateCert(cluster); err != nil {
		return err
	}

	return certphases.SendCert(cluster)
}
