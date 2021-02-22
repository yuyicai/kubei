package options

import (
	flag "github.com/spf13/pflag"
	"github.com/yuyicai/kubei/internal/constants"
)

const (
	Key                       = "key"
	ShortKey                  = "k"
	Port                      = "port"
	Password                  = "password"
	ShortPassword             = "p"
	User                      = "user"
	KubernetesVersion         = "kubernetes-version"
	ContainerEngineVersion    = "container-engine-version"
	ControlPlaneEndpoint      = "control-plane-endpoint"
	ImageRepository           = "image-repository"
	Masters                   = "masters"
	ShortMasters              = "m"
	Workers                   = "nodes"
	ShortNodes                = "n"
	PodNetworkCidr            = "pod-network-cidr"
	ServiceCidr               = "service-cidr"
	JumpServer                = "jump-server"
	RemoveContainerEngine     = "remove-container-engine"
	RemoveKubernetesComponent = "remove-kubernetes-component"
	OfflineFile               = "offline-file"
	ShortOfflineFile          = "f"
	CertNotAfterTime          = "cert-time"
	NetworkPlugin             = "network-plugin"
	Online                    = "install-online"
)

func AddResetFlags(flagSet *flag.FlagSet, options *Reset) {
	flagSet.BoolVar(
		&options.RemoveContainerEngine, RemoveContainerEngine, options.RemoveContainerEngine,
		"If true, remove the container engine from the nodes",
	)

	flagSet.BoolVar(
		&options.RemoveKubeComponent, RemoveKubernetesComponent, options.RemoveKubeComponent,
		"If true, remove the kubernetes component from the nodes",
	)
}

func AddContainerEngineConfigFlags(flagSet *flag.FlagSet, options *ContainerEngine) {
	flagSet.StringVar(
		&options.Version, ContainerEngineVersion, options.Version,
		"The Docker version.",
	)
}

func AddKubeClusterNodesConfigFlags(flagSet *flag.FlagSet, options *ClusterNodes) {
	flagSet.StringSliceVarP(
		&options.Masters, Masters, ShortMasters, options.Masters,
		"The master nodes IP",
	)

	flagSet.StringSliceVarP(
		&options.Workers, Workers, ShortNodes, options.Workers,
		"The worker nodes IP",
	)
}

func AddPublicUserInfoConfigFlags(flagSet *flag.FlagSet, options *PublicHostInfo) {
	flagSet.StringVar(
		&options.User, User, constants.DefaultSSHUser,
		"SSH user of the nodes.",
	)

	flagSet.StringVarP(
		&options.Password, Password, ShortPassword, options.Password,
		"SSH password of the nodes.",
	)

	flagSet.StringVar(
		&options.Port, Port, constants.DefaultSSHPort,
		"SSH port of the nodes.",
	)

	flagSet.StringVarP(
		&options.Key, Key, ShortKey, options.Key,
		"SSH key of the nodes.",
	)
}

func AddKubeadmConfigFlags(flagSet *flag.FlagSet, options *Kubeadm) {
	flagSet.StringVar(
		&options.Networking.ServiceSubnet, ServiceCidr, constants.DefaultServiceSubnet,
		"Use alternative range of IP address for service VIPs",
	)
	flagSet.StringVar(
		&options.Networking.PodSubnet, PodNetworkCidr, constants.DefaultPodNetworkCidr,
		"Specify range of IP addresses for the pod network",
	)

	AddImageMetaFlags(flagSet, &options.ImageRepository)
	AddControlPlaneEndpointFlags(flagSet, options)
}

func AddControlPlaneEndpointFlags(flagSet *flag.FlagSet, options *Kubeadm) {
	flagSet.StringVar(
		&options.ControlPlaneEndpoint, ControlPlaneEndpoint, constants.DefaultControlPlaneEndpoint,
		`Specify a DNS name for the control plane.`,
	)
}

func AddImageMetaFlags(flagSet *flag.FlagSet, imageRepository *string) {
	flagSet.StringVar(imageRepository, ImageRepository, constants.DefaultImageRepository,
		"Choose a container registry to pull control plane images from",
	)
}

func AddJumpServerFlags(flagSet *flag.FlagSet, userInfo *map[string]string) {
	flagSet.StringToStringVar(userInfo, JumpServer, *userInfo,
		"Jump server user info",
	)
}

func AddOfflinePackageFlags(flagSet *flag.FlagSet, pkg *string) {
	flagSet.StringVarP(pkg, OfflineFile, ShortOfflineFile, *pkg,
		"Path to offline file path",
	)
}

func AddCertNotAfterTimeFlags(flagSet *flag.FlagSet, year *int) {
	flagSet.IntVar(year, CertNotAfterTime, constants.DefaultCertNotAfterYear,
		"cert not after time, time units is year",
	)
}

func AddNetworkPluginFlags(flagSet *flag.FlagSet, networkType *string) {
	flagSet.StringVar(networkType, NetworkPlugin, constants.DefaulNetworkPlugin,
		"network plugin",
	)
}

func AddOnlineFlags(flagSet *flag.FlagSet, options *bool) {
	flagSet.BoolVar(
		options, Online, *options,
		"If true, install kubernetes cluster online",
	)
}
		
func AddKubernetesFlags(flagSet *flag.FlagSet, options *Kubernetes) {
	flagSet.StringVar(
		&options.Version, KubernetesVersion, options.Version,
		"The Kubernetes version",
	)
}


