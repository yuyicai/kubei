package rundata

import "github.com/yuyicai/kubei/pkg/ssh"

type ClusterNodes struct {
	Masters []*Node
	Worker  []*Node
}

func (c *ClusterNodes) GetAllMastersHost() []string {
	var hosts []string
	for _, master := range c.Masters {
		hosts = append(hosts, master.HostInfo.Host)
	}
	return hosts
}

func (c *ClusterNodes) GetAllNodes() []*Node {
	return append(c.Masters, c.Worker...)
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
