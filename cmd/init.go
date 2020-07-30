package cmd

import (
	"io"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	initphases "github.com/yuyicai/kubei/cmd/phases/init"
	"github.com/yuyicai/kubei/config/options"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/preflight"
)

// NewCmdInit returns "kubei init" command.
func NewCmdInit(out io.Writer, initOptions *runOptions) *cobra.Command {
	if initOptions == nil {
		initOptions = newInitOptions()
	}
	initRunner := workflow.NewRunner()

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Run this command in order to create a high availability Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := initRunner.InitData(args)
			if err != nil {
				return err
			}

			data := c.(*runData)
			cluster := data.Cluster()
			if err := preflight.Prepare(cluster); err != nil {
				return err
			}

			if err := initRunner.Run(args); err != nil {
				return err
			}

			preflight.CloseSSH(cluster)

			return nil
		},
		Args: cobra.NoArgs,
	}

	// adds flags to the init command
	// init command local flags could be eventually inherited by the sub-commands automatically generated for phases
	addInitConfigFlags(cmd.Flags(), initOptions.kubei)
	options.AddKubeadmConfigFlags(cmd.Flags(), initOptions.kubeadm)

	// initialize the workflow runner with the list of phases
	initRunner.AppendPhase(initphases.NewSendPhase())
	initRunner.AppendPhase(initphases.NewContainerEnginePhase())
	initRunner.AppendPhase(initphases.NewKubeComponentPhase())
	initRunner.AppendPhase(initphases.NewCertPhase())
	initRunner.AppendPhase(initphases.NewKubeadmPhase())

	// sets the rundata builder function, that will be used by the runner
	// both when running the entire workflow or single phases
	initRunner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return newInitData(cmd, args, initOptions, out)
	})

	// binds the Runner to kubei init command by altering
	// command help, adding --skip-phases flag and by adding phases subcommands
	initRunner.BindToCommand(cmd)

	return cmd
}

func addInitConfigFlags(flagSet *flag.FlagSet, k *options.Kubei) {
	options.AddContainerEngineConfigFlags(flagSet, &k.ContainerEngine)
	options.AddPublicUserInfoConfigFlags(flagSet, &k.ClusterNodes.PublicHostInfo)
	options.AddKubeClusterNodesConfigFlags(flagSet, &k.ClusterNodes)
	options.AddJumpServerFlags(flagSet, &k.JumpServer)
	options.AddOfflinePackageFlags(flagSet, &k.OfflineFile)
}

func newInitOptions() *runOptions {
	kubeiOptions := options.NewKubei()
	kubeadmOptions := options.NewKubeadm()

	return &runOptions{
		kubei:   kubeiOptions,
		kubeadm: kubeadmOptions,
	}
}

func newInitData(cmd *cobra.Command, args []string, options *runOptions, out io.Writer) (*runData, error) {

	clusterCfg := rundata.NewCluster()

	options.kubei.ApplyTo(clusterCfg.Kubei)
	options.kubeadm.ApplyTo(clusterCfg.Kubeadm)

	rundata.DefaultKubeiCfg(clusterCfg.Kubei)
	rundata.DefaultkubeadmCfg(clusterCfg.Kubeadm, clusterCfg.Kubei)

	initDatacfg := &runData{
		cluster: clusterCfg,
	}

	return initDatacfg, nil
}
