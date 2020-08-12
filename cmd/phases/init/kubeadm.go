package init

import (
	"context"
	"errors"

	"github.com/fatih/color"
	"github.com/go-kratos/kratos/pkg/sync/errgroup"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"github.com/yuyicai/kubei/cmd/phases"
	"github.com/yuyicai/kubei/config/options"
	kubeadmphases "github.com/yuyicai/kubei/phases/kubeadm"
	networkphases "github.com/yuyicai/kubei/phases/network"
)

// NewKubeadmPhase creates a kubei workflow phase that implements handling of kubeadm.
func NewKubeadmPhase() workflow.Phase {
	phase := workflow.Phase{
		Name:         "kubeadm",
		Short:        "create a k8s cluster with kubeadm",
		Long:         "create a k8s cluster with kubeadm",
		InheritFlags: getKubeadmPhaseFlags(),
		Run:          runKubeadm,
	}
	return phase
}

func getKubeadmPhaseFlags() []string {
	flags := []string{
		options.OfflineFile,
		options.JumpServer,
		options.ControlPlaneEndpoint,
		options.ImageRepository,
		options.PodNetworkCidr,
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

func runKubeadm(c workflow.RunData) error {
	data, ok := c.(phases.RunData)
	if !ok {
		return errors.New("kubeadm phase invoked with an invalid rundata struct")
	}

	cluster := data.Cluster()

	if err := kubeadmphases.LoadOfflineImages(cluster); err != nil {
		return err
	}

	color.HiBlue("Creating Kubernetes cluster ☸️")
	// init master0
	if err := kubeadmphases.InitMaster(cluster); err != nil {
		return err
	}

	// add network plugin
	color.HiBlue("Installing Network plugin 🌐")
	if err := networkphases.Network(cluster); err != nil {
		return err
	}

	g := errgroup.WithCancel(context.Background())
	// join to master nodes
	g.Go(func(ctx context.Context) error {
		return kubeadmphases.JoinControlPlane(cluster)
	})
	// join to worker nodes
	// and set ha
	g.Go(func(ctx context.Context) error {
		return kubeadmphases.JoinNode(cluster)
	})
	if err := g.Wait(); err != nil {
		return err
	}

	return kubeadmphases.CheckNodesReady(cluster)
}
