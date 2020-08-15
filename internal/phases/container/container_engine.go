package container

import (
	"fmt"
	"github.com/yuyicai/kubei/internal/constants"
	"github.com/yuyicai/kubei/internal/rundata"
)

func InstallContainerEngine(c *rundata.Cluster) error {
	switch c.ContainerEngine.Type {
	case constants.ContainerEngineTypeDocker:
		return InstallDocker(c)
	case constants.ContainerEngineTypeContainerd:
		//TODO
	case constants.ContainerEngineTypeCRIO:
		//TODO
	default:
		fmt.Println("Uninstall container Engine")
	}
	return nil
}
