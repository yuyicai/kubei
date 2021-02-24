package rundata

import (
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

// +k8s:deepcopy-gen=false

type Node struct {
	SSH                   *ssh.Client
	HostInfo              HostInfo
	CertificateTree       CertificateTree
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
