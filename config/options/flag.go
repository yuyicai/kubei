package options

import (
	flag "github.com/spf13/pflag"
)

const (
	Key                  = "key"
	Port                 = "port"
	Password             = "password"
	User                 = "user"
	KubernetesVersion    = "kubernetes-version"
	DockerVersion        = "docker-version"
	ControlPlaneEndpoint = "control-plane-endpoint"
	ImageRepository      = "image-repository"
	Masters              = "masters"
	Workers              = "workers"
	PodNetworkCidr       = "pod-network-cidr"
	ServiceCidr          = "service-cidr"
	JumpServer           = "jump-server"
)

func AddcontainerEngineConfigFlags(flagSet *flag.FlagSet, options *containerEngine) {
	flagSet.StringVar(
		&options.Version, DockerVersion, options.Version,
		"The Docker version.",
	)
}

func AddKubeComponentConfigFlags(flagSet *flag.FlagSet, options *KubeComponent) {
	flagSet.StringVar(
		&options.Version, KubernetesVersion, options.Version,
		"The Kubernetes version",
	)
}

func AddKubeClusterNodesConfigFlags(flagSet *flag.FlagSet, options *ClusterNodes) {
	flagSet.StringSliceVar(
		&options.Masters, Masters, options.Masters,
		"The master nodes IP",
	)

	flagSet.StringSliceVar(
		&options.Workers, Workers, options.Workers,
		"The worker nodes IP",
	)
}

func AddPublicUserInfoConfigFlags(flagSet *flag.FlagSet, options *PublicHostInfo) {
	flagSet.StringVar(
		&options.User, User, "root",
		"SSH user of the nodes.",
	)

	flagSet.StringVar(
		&options.Password, Password, options.Password,
		"SSH password of the nodes.",
	)

	flagSet.StringVar(
		&options.Port, Port, "22",
		"SSH port of the nodes.",
	)

	flagSet.StringVar(
		&options.Key, Key, options.Key,
		"SSH key of the nodes.",
	)
}

func AddKubeadmConfigFlags(flagSet *flag.FlagSet, options *Kubeadm) {
	flagSet.StringVar(
		&options.ControlPlaneEndpoint, ControlPlaneEndpoint, "apiserver.k8s.local:6443",
		`Specify a DNS name for the control plane.`,
	)

	flagSet.StringVar(
		&options.Networking.ServiceSubnet, ServiceCidr, "10.96.0.0/12",
		"Use alternative range of IP address for service VIPs.",
	)
	flagSet.StringVar(
		&options.Networking.PodSubnet, PodNetworkCidr, "10.244.0.0/16",
		"Specify range of IP addresses for the pod network. If set, the control plane will automatically allocate CIDRs for every node.",
	)

	AddImageMetaFlags(flagSet, &options.ImageRepository)
}

func AddImageMetaFlags(flagSet *flag.FlagSet, imageRepository *string) {
	flagSet.StringVar(imageRepository, ImageRepository, "gcr.azk8s.cn/google_containers",
		"Choose a container registry to pull control plane images from",
	)
}

func AddJumpServerFlags(flagSet *flag.FlagSet, userInfo *map[string]string) {
	flagSet.StringToStringVar(userInfo, JumpServer, *userInfo,
		"Jump server user info, apply with \"--jump-server host=IP,port=22,user=your-user,password=your-password,key=key-path\"",
	)
}
