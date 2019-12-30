package options

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/yuyicai/kubei/config/rundata"
	"k8s.io/klog"
)

func (c *ClusterNodes) ApplyTo(data *rundata.ClusterNodes) {

	setNodesHost(c.Masters, &data.Masters)
	setNodesHost(c.Workers, &data.Worker)

	nodes := append(data.Masters, data.Worker...)

	for _, v := range nodes {
		if v.HostInfo.Password == "" && c.PublicHostInfo.Password != "" {
			v.HostInfo.Password = c.PublicHostInfo.Password
		}
		if v.HostInfo.User == "" && c.PublicHostInfo.User != "" {
			v.HostInfo.User = c.PublicHostInfo.User
		}
		if v.HostInfo.Port == "" && c.PublicHostInfo.Port != "" {
			v.HostInfo.Port = c.PublicHostInfo.Port
		}
		if v.HostInfo.Key == "" && c.PublicHostInfo.Key != "" {
			v.HostInfo.Key = c.PublicHostInfo.Key
		}
		if v.Name == "" {
			v.Name = v.HostInfo.Host
		}
	}
}

func (c *ContainerEngine) ApplyTo(data *rundata.ContainerEngine) {
	if c.Version != "" {
		data.Version = strings.Replace(c.Version, "v", "", -1)
	}
}

func (c *Reset) ApplyTo(data *rundata.Reset) {
	if c.RemoveKubeComponent {
		data.RemoveKubeComponent = c.RemoveKubeComponent
	}

	if c.RemoveContainerEngine {
		data.RemoveContainerEngine = c.RemoveContainerEngine
	}
}

func (k *Kubei) ApplyTo(data *rundata.Kubei) {

	k.ContainerEngine.ApplyTo(&data.ContainerEngine)
	k.ClusterNodes.ApplyTo(&data.ClusterNodes)
	k.Reset.ApplyTo(&data.Reset)

	if len(k.JumpServer) > 0 {
		data.JumpServer.IsUse = true

		if err := mapstructure.Decode(k.JumpServer, &data.JumpServer.HostInfo); err != nil {
			klog.Fatal(err)
		}
	}
}

func setNodesHost(optionsNodes []string, nodes *[]*rundata.Node) {
	if len(optionsNodes) > 0 {
		for _, v := range optionsNodes {
			v = strings.Replace(v, " ", "", -1)
			vv := strings.Split(v, ";")
			node := &rundata.Node{}
			node.HostInfo.Host = vv[0]
			if len(vv) > 1 {
				//TODO set nodes ssh host info (host,user,port,password,key) with --masters and --workers
			}
			*nodes = append(*nodes, node)
		}
	}
}
