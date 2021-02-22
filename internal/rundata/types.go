package rundata

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/go-kratos/kratos/pkg/sync/errgroup"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"github.com/yuyicai/kubei/internal/constants"
	"github.com/yuyicai/kubei/pkg/ssh"
)

type Cluster struct {
	*Kubei
	Kubeadm *Kubeadm
	Mutex   sync.Mutex
}

func (c *Cluster) String() string {
	return fmt.Sprintf("%+v\n%+v", c.Kubei, c.Kubeadm)
}

func (c *Cluster) RunOnAllNodes(f func(*Node) error) error {
	return run(c.ClusterNodes.GetAllNodes(), f)
}

func (c *Cluster) RunOnMasters(f func(*Node) error) error {
	return run(c.ClusterNodes.Masters, f)
}

func (c *Cluster) RunOnWorkers(f func(*Node) error) error {
	return run(c.ClusterNodes.Workers, f)
}

func (c *Cluster) RunOnWorkersAndPrintLog(f func(*Node) error, s string) error {
	if len(c.ClusterNodes.Workers) == 0 {
		return nil
	}
	fmt.Println(s)
	return run(c.ClusterNodes.Workers, f)
}

func (c *Cluster) RunOnOtherMastersAndPrintLog(f func(*Node) error, s string) error {
	if len(c.ClusterNodes.Masters) <= 1 {
		return nil
	}
	fmt.Println(s)
	return run(c.ClusterNodes.Masters[1:], f)
}

func (c *Cluster) RunOnOtherMasters(f func(*Node) error) error {
	if len(c.ClusterNodes.Masters) <= 1 {
		return nil
	}

	return run(c.ClusterNodes.Masters[1:], f)
}

func (c *Cluster) RunOnOtherMastersOneByOne(f func(*Node) error) error {
	if len(c.ClusterNodes.Masters) <= 1 {
		return nil
	}

	for _, node := range c.ClusterNodes.Masters[1:] {
		if err := runOne(node, f); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cluster) RunOnFirstMaster(f func(*Node) error) error {
	if len(c.ClusterNodes.Masters) == 0 {
		return errors.New("not master")
	}

	return runOne(c.ClusterNodes.Masters[0], f)
}

func run(nodes []*Node, f func(*Node) error) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range nodes {
		node := node
		g.Go(func(ctx context.Context) error {
			return runOne(node, f)
		})
	}

	return g.Wait()
}

func runOne(node *Node, f func(*Node) error) error {
	return f(node)
}

type Kubei struct {
	ContainerEngine  ContainerEngine
	Kubernetes       Kubernetes
	ClusterNodes     ClusterNodes
	NetworkPlugins   NetworkPlugins
	HA               HA
	JumpServer       JumpServer
	Install          Install
	Reset            Reset
	Addons           Addons
	OfflineFile      string
	Online           bool
	CertNotAfterTime int
}

type JumpServer struct {
	*ssh.Client
	HostInfo HostInfo
}

type Reset struct {
	RemoveContainerEngine bool
	RemoveKubeComponent   bool
}

type Install struct {
	Type string
}

type Kubeadm struct {
	kubeadmapi.InitConfiguration
}

func NewKubei() *Kubei {
	return &Kubei{
		ContainerEngine: ContainerEngine{},
		Kubernetes:      Kubernetes{},
		ClusterNodes:    ClusterNodes{},
		JumpServer:      JumpServer{},
		Install:         Install{},
		Reset:           Reset{},
		Addons:          Addons{},
	}
}

func NewKubeadm() *Kubeadm {
	return &Kubeadm{
		InitConfiguration: kubeadmapi.InitConfiguration{},
	}
}

func NewCluster() *Cluster {
	return &Cluster{
		Kubei: &Kubei{
			ContainerEngine: ContainerEngine{},
			Kubernetes:      Kubernetes{},
			ClusterNodes:    ClusterNodes{},
			JumpServer:      JumpServer{},
			Install:         Install{},
			Reset:           Reset{},
			Addons:          Addons{},
		},
		Kubeadm: &Kubeadm{
			InitConfiguration: kubeadmapi.InitConfiguration{},
		},
	}
}
