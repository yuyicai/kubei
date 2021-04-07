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
	DefaultCGroupDriver           = "cgroupfs"
	DefaultLogDriver              = "json-file"
	DefaultLogOptsMaxSize         = "500m"
	DockerDefaultStorageDriver    = "overlay2"

	// kubeadm
	DefaultServiceSubnet        = "10.96.0.0/12"
	DefaultPodNetworkCidr       = "10.244.0.0/16"
	DefaultControlPlaneEndpoint = "apiserver.k8s.local:6443"
	DefaultImageRepository      = "k8s.gcr.io"
	DefaultAPIBindPort          = 6443
	DefaultClusterName          = "kubernetes"
	DefaultWaitNodeInterval     = 2 * time.Second
	DefaultWaitNodeTimeout      = 6 * time.Minute
	DefaultCertNotAfterYear     = 10
	DefaultCertNotAfterTime     = Year * DefaultCertNotAfterYear

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
	DefaultNginxVersion         = "1.19.9"
	DefaultNginxPort            = "6443"

	LoopbackAddress = "127.0.0.1"

	DefaultGOMAXPROCS = 20

	Day  = 24 * time.Hour
	Year = 365 * Day
)
