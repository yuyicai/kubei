package constants

import "time"

const (
	DefaultSSHUser = "root"
	DefaultSSHPort = "22"

	DefaultServiceSubnet        = "10.96.0.0/12"
	DefaultPodNetworkCidr       = "10.244.0.0/16"
	DefaultControlPlaneEndpoint = "apiserver.k8s.local:6443"
	DefaultImageRepository      = "gcr.azk8s.cn/google_containers"
	DefaultAPIBindPort          = "6443"
	DefaulNetworkPlugin         = "flannel"

	DefaultFlannelImageRepository = "quay.azk8s.cn/coreos"
	DefaultFlannelImageName       = "flannel"
	DefaultFlannelVersion         = "v0.11.0-amd64"
	DefaultFlannelBackendType     = "vxlan"

	DefaultNginxImageRepository = ""
	DefaultNginxImageName       = "nginx"
	DefaultNginxVersion         = "1.17"
	DefaultNginxPort            = "6443"

	LoopbackAddress = "127.0.0.1"

	DefaultLocalSLBInterval = 2 * time.Second
	DefaultLocalSLBTimeout  = 6 * time.Minute

	HATypeNone = 1 << iota
	HATypeLocalSLB
	HATypeExternalSLB

	LocalSLBTypeNginx = 1 << iota
	LocalSLBTypeHAproxy

	InstallationTypeOffline = 1 << iota
	InstallationTypeApt
	InstallationTypeYum
)
