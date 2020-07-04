package container

import (
	"fmt"
	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
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
