package phases

import (
	"context"
	"errors"
	"github.com/bilibili/kratos/pkg/sync/errgroup"
	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/options"
	kubeadmphases "github.com/yuyicai/kubei/phases/kubeadm"
	networkphases "github.com/yuyicai/kubei/phases/network"
	"github.com/yuyicai/kubei/preflight"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
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
	data, ok := c.(InitData)
	if !ok {
		return errors.New("kubeadm phase invoked with an invalid rundata struct")
	}

	kubeiCfg := data.KubeiCfg()
	kubeadmCfg := data.KubeadmCfg()

	if len(kubeiCfg.ClusterNodes.GetAllMastersHost()) == 0 {
		return errors.New("You host to set the master nodes by the flag --masters")
	}

	if err := preflight.Check(kubeiCfg); err != nil {
		return err
	}

	// init master0
	masters := kubeiCfg.ClusterNodes.Masters
	masters0 := masters[0]
	if err := kubeadmphases.InitMaster(masters0, kubeiCfg, kubeadmCfg); err != nil {
		return err
	}

	// add network plugin
	if err := networkphases.Network(masters0, kubeiCfg.NetworkPlugins, kubeadmCfg.Networking); err != nil {
		return err
	}

	g := errgroup.WithCancel(context.Background())

	// join to master nodes
	if len(masters) > 1 {
		h := &kubeiCfg.HA.Type
		if *h == constants.HATypeNone {
			*h = constants.HATypeLocalSLB
		}

		g.Go(func(ctx context.Context) error {
			return kubeadmphases.JoinControlPlane(masters, kubeiCfg, kubeadmCfg)
		})
	}

	// join to worker nodes
	// and set ha
	g.Go(func(ctx context.Context) error {
		return kubeadmphases.JoinNode(kubeiCfg, kubeadmCfg)
	})

	if err := g.Wait(); err != nil {
		return err
	}

	// waiting for all nodes to become ready
	output, done := kubeadmphases.CheckNodesReady(masters[0], constants.DefaultWaitNodeInterval, constants.DefaultWaitNodeTimeout)
	if done {
		klog.Info(output, "\nKubernetes High-Availability cluster deployment completed")
	}

	return nil
}
