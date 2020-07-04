package rundata

import (
	"context"
	"errors"
	"github.com/bilibili/kratos/pkg/sync/errgroup"
	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/pkg/ssh"
)

type ClusterNodes struct {
	Masters []*Node
	Workers []*Node
}

func (c *ClusterNodes) GetAllMastersHost() []string {
	var hosts []string
	for _, master := range c.Masters {
		hosts = append(hosts, master.HostInfo.Host)
	}
	return hosts
}

func (c *ClusterNodes) GetAllNodes() []*Node {
	return append(c.Masters, c.Workers...)
}

func (c *ClusterNodes) Run(cmd string) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range c.GetAllNodes() {
		node := node
		g.Go(func(ctx context.Context) error {
			return node.Run(cmd)
		})
	}

	return g.Wait()
}

func (c *ClusterNodes) MastersRun(cmd string) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range c.Masters {
		node := node
		g.Go(func(ctx context.Context) error {
			return node.Run(cmd)
		})
	}

	return g.Wait()
}

func (c *ClusterNodes) FirstMasterRun(cmd string) error {
	if len(c.Masters) == 0 {
		return errors.New("not master")
	}
	return c.Masters[0].Run(cmd)
}

func (c *ClusterNodes) WorkersRun(cmd string) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range c.Workers {
		node := node
		g.Go(func(ctx context.Context) error {
			return node.Run(cmd)
		})
	}

	return g.Wait()
}

func (c *ClusterNodes) LogRun(f func(node *Node) error) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range c.Workers {
		node := node
		g.Go(func(ctx context.Context) error {
			return f(node)
		})
	}

	return g.Wait()
}

type Node struct {
	SSH                   *ssh.Client
	HostInfo              HostInfo
	Name                  string
	PackageManagementType string
	InstallType           string
	IsSend                bool
}

type HostInfo struct {
	Host     string
	User     string
	Password string
	Port     string
	Key      string
}

func (n *Node) Run(cmd string) error {
	return n.SSH.Run(cmd)
}

func (n *Node) RunOut(cmd string) ([]byte, error) {
	return n.SSH.RunOut(cmd)
}
