package operator

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/pkg/sync/errgroup"
	"github.com/pkg/errors"
	"github.com/yuyicai/kubei/internal/constants"
	"github.com/yuyicai/kubei/internal/rundata"
)

type Tasks func(*rundata.Node, *rundata.Cluster) error

func RunOnAllNodes(c *rundata.Cluster, tasks Tasks) error {
	return run(c.ClusterNodes.GetAllNodes(), c, tasks)
}

func RunOnMasters(c *rundata.Cluster, tasks Tasks) error {
	return run(c.ClusterNodes.Masters, c, tasks)
}

func RunOnWorkers(c *rundata.Cluster, tasks Tasks) error {
	return run(c.ClusterNodes.Workers, c, tasks)
}

func RunOnWorkersWithMsg(c *rundata.Cluster, tasks Tasks, s string) error {
	if len(c.ClusterNodes.Workers) == 0 {
		return nil
	}
	fmt.Println(s)
	return run(c.ClusterNodes.Workers, c, tasks)
}

func RunOnOtherMastersWithMsg(c *rundata.Cluster, tasks Tasks, s string) error {
	if len(c.ClusterNodes.Masters) <= 1 {
		return nil
	}
	fmt.Println(s)
	return run(c.ClusterNodes.Masters[1:], c, tasks)
}

func RunOnOtherMasters(c *rundata.Cluster, tasks Tasks) error {
	if len(c.ClusterNodes.Masters) <= 1 {
		return nil
	}

	return run(c.ClusterNodes.Masters[1:], c, tasks)
}

func RunOnOtherMastersOneByOne(c *rundata.Cluster, tasks Tasks) error {
	if len(c.ClusterNodes.Masters) <= 1 {
		return nil
	}

	for _, node := range c.ClusterNodes.Masters[1:] {
		if err := runOne(node, c, tasks); err != nil {
			return err
		}
	}
	return nil
}

func RunOnFirstMaster(c *rundata.Cluster, tasks Tasks) error {
	if len(c.ClusterNodes.Masters) == 0 {
		return errors.New("not master")
	}

	return runOne(c.ClusterNodes.Masters[0], c, tasks)
}

func run(nodes []*rundata.Node, c *rundata.Cluster, f Tasks) error {
	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(constants.DefaultGOMAXPROCS)
	for _, node := range nodes {
		node := node
		g.Go(func(ctx context.Context) error {
			return runOne(node, c, f)
		})
	}

	return g.Wait()
}

func runOne(node *rundata.Node, c *rundata.Cluster, tasks Tasks) error {
	return tasks(node, c)
}
