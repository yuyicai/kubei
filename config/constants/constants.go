package constants

import "time"

const (
	// ssh
	DefaultSSHUser = "root"
	DefaultSSHPort = "22"

	InstallTypeOffline       = "offline"
	InstallTypeOnline        = "online"
	PackageManagementTypeApt = "apt"
	PackageManagementTypeYum = "yum"
	DefaultLocalSLBInterval  = 2 * time.Second
	DefaultLocalSLBTimeout   = 6 * time.Minute

	// container engine
	ContainerEngineTypeDocker     = "docker"
	ContainerEngineTypeContainerd = "containerd"
	ContainerEngineTypeCRIO       = "cri-o"
	RegistryMirrors               = "https://dockerhub.mirrors.nwafu.edu.cn/"
	DefaultCGroupDriver           = "systemd"
	DefaultLogDriver              = "json-file"
	DefaultLogOptsMaxSize         = "500m"
	DockerDefaultStorageDriver    = "overlay2"

	// kubeadm
	DefaultServiceSubnet        = "10.96.0.0/12"
	DefaultPodNetworkCidr       = "10.244.0.0/16"
	DefaultControlPlaneEndpoint = "apiserver.k8s.local:6443"
	DefaultImageRepository      = "gcr.azk8s.cn/google_containers"
	DefaultAPIBindPort          = 6443
	DefaultWaitNodeInterval     = 2 * time.Second
	DefaultWaitNodeTimeout      = 6 * time.Minute

	// networking plugin
	DefaulNetworkPlugin           = "flannel"
	DefaultFlannelImageRepository = "quay.io/coreos"
	DefaultFlannelImageName       = "flannel"
	DefaultFlannelVersion         = "v0.11.0-amd64"
	DefaultFlannelBackendType     = "vxlan"

	// ha
	LocalSLBTypeNginx           = "nginx"
	LocalSLBTypeHAproxy         = "haproxy"
	HATypeNone                  = "none"
	HATypeLocalSLB              = "local"
	HATypeExternalSLB           = "external"
	DefaultNginxImageRepository = ""
	DefaultNginxImageName       = "nginx"
	DefaultNginxVersion         = "1.17"
	DefaultNginxPort            = "6443"

	LoopbackAddress = "127.0.0.1"

	DefaultGOMAXPROCS = 20
)
