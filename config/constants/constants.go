package constants

import "time"

const (
	// ssh
	DefaultSSHUser = "root"
	DefaultSSHPort = "22"

	// kubeadm
	DefaultServiceSubnet        = "10.96.0.0/12"
	DefaultPodNetworkCidr       = "10.244.0.0/16"
	DefaultControlPlaneEndpoint = "apiserver.k8s.local:6443"
	DefaultImageRepository      = "gcr.azk8s.cn/google_containers"
	DefaultAPIBindPort          = 6443

	// networking plugin
	DefaulNetworkPlugin = "flannel"

	// flannel
	DefaultFlannelImageRepository = "quay.azk8s.cn/coreos"
	DefaultFlannelImageName       = "flannel"
	DefaultFlannelVersion         = "v0.11.0-amd64"
	DefaultFlannelBackendType     = "vxlan"

	// nginx
	DefaultNginxImageRepository = ""
	DefaultNginxImageName       = "nginx"
	DefaultNginxVersion         = "1.17"
	DefaultNginxPort            = "6443"

	LoopbackAddress = "127.0.0.1"

	DefaultLocalSLBInterval = 2 * time.Second
	DefaultLocalSLBTimeout  = 6 * time.Minute

	IsNotSet = 0
)

const (
	HATypeNone = 1 << iota
	HATypeLocalSLB
	HATypeExternalSLB
)

const (
	LocalSLBTypeNginx = 1 << iota
	LocalSLBTypeHAproxy
)

const (
	InstallationTypeOffline = 1 << iota
	InstallationTypeApt
	InstallationTypeYum
)

// container engine
const (
	ContainerEngineTypeDocker = 1 << iota
	ContainerEngineTypeContainerd
	ContainerEngineTypeCRIO

	RegistryMirrors       = "https://dockerhub.azk8s.cn"
	DefaultCGroupDriver   = "systemd"
	DefaultLogDriver      = "json-file"
	DefaultLogOptsMaxSize = "500m"

	DockerDefaultStorageDriver = "overlay2"
)
