package options

import (
	"github.com/yuyicai/kubei/internal/constants"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/yuyicai/kubei/internal/rundata"
	"k8s.io/klog"
)

func (c *ClusterNodes) ApplyTo(data *rundata.ClusterNodes) {

	setNodesHost(&data.Masters, c.Masters)
	setNodesHost(&data.Workers, c.Workers)
	nodes := append(data.Masters, data.Workers...)

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
		data.Docker.Version = strings.Replace(c.Version, "v", "", -1)
	}
}

func (k *Kubernetes) ApplyTo(data *rundata.Kubernetes) {
	if k.Version != "" {
		data.Version = strings.Replace(k.Version, "v", "", -1)
	}
}

func (r *Reset) ApplyTo(data *rundata.Reset) {
	if r.RemoveKubeComponent {
		data.RemoveKubeComponent = r.RemoveKubeComponent
	}

	if r.RemoveContainerEngine {
		data.RemoveContainerEngine = r.RemoveContainerEngine
	}
}

func (k *Kubei) ApplyTo(data *rundata.Kubei) {

	k.ContainerEngine.ApplyTo(&data.ContainerEngine)
	k.ClusterNodes.ApplyTo(&data.ClusterNodes)
	k.Reset.ApplyTo(&data.Reset)

	if len(k.JumpServer) > 0 {
		if err := mapstructure.Decode(k.JumpServer, &data.JumpServer.HostInfo); err != nil {
			klog.Fatal(err)
		}
	}

	if k.OfflineFile != "" {
		data.OfflineFile = k.OfflineFile
		setNodesInstallType(data.ClusterNodes.GetAllNodes())
	}

	data.NetworkPlugins.Type = k.NetworkType

	data.CertNotAfterTime = k.CertNotAfterTime
}

func setNodesHost(nodes *[]*rundata.Node, optionsNodes []string) {
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

func setNodesInstallType(nodes []*rundata.Node) {
	for _, node := range nodes {
		node.InstallType = constants.InstallTypeOffline
	}
}
