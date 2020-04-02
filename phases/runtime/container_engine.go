package runtime

import (
	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
)

func InstallContainerEngine(nodes []*rundata.Node, c rundata.ContainerEngine) error {
	switch c.Type {
	case constants.ContainerEngineTypeDocker:
		return InstallDocker(nodes, c.Docker)
	case constants.ContainerEngineTypeContainerd:
		//TODO
	case constants.ContainerEngineTypeCRIO:
		//TODO
	}
	return nil
}
