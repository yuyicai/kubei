package runtime

import "github.com/yuyicai/kubei/config/rundata"

func InstallContainerEngine(nodes []*rundata.Node, c rundata.ContainerEngine) error {
	if err := InstallDocker(nodes, c.Docker); err != nil {
		return err
	}
	return nil
}
