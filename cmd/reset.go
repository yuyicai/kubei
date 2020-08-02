package cmd

import (
	"github.com/yuyicai/kubei/preflight"
	"io"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	phases "github.com/yuyicai/kubei/cmd/phases/reset"
	"github.com/yuyicai/kubei/config/options"
	"github.com/yuyicai/kubei/config/rundata"
)

// NewCmdreset returns "kubei reset" command.
func NewCmdReset(out io.Writer, runOptions *runOptions) *cobra.Command {
	if runOptions == nil {
		runOptions = newResetOptions()
	}
	resetRunner := workflow.NewRunner()
	cluster := &rundata.Cluster{}

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Run this command in order to reset nodes",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			c, err := resetRunner.InitData(args)
			if err != nil {
				return err
			}

			data := c.(*runData)
			cluster = data.Cluster()
			return preflight.Prepare(cluster)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return resetRunner.Run(args)
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			preflight.CloseSSH(cluster)
			return nil
		},
		Args: cobra.NoArgs,
	}

	// adds flags to the reset command
	// reset command local flags could be eventually inherited by the sub-commands automatically generated for phases
	addResetConfigFlags(cmd.Flags(), runOptions.kubei)
	options.AddControlPlaneEndpointFlags(cmd.Flags(), runOptions.kubeadm)

	// initialize the workflow runner with the list of phases
	resetRunner.AppendPhase(phases.NewKubeadmPhase())
	resetRunner.AppendPhase(phases.NewKubeComponentPhase())
	resetRunner.AppendPhase(phases.NewContainerEnginePhase())

	// sets the rundata builder function, that will be used by the runner
	// both when running the entire workflow or single phases
	resetRunner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return newResetData(cmd, args, runOptions, out)
	})

	// binds the Runner to kubei reset command by altering
	// command help, adding --skip-phases flag and by adding phases subcommands
	resetRunner.BindToCommand(cmd)

	return cmd
}

func addResetConfigFlags(flagSet *flag.FlagSet, k *options.Kubei) {
	options.AddPublicUserInfoConfigFlags(flagSet, &k.ClusterNodes.PublicHostInfo)
	options.AddKubeClusterNodesConfigFlags(flagSet, &k.ClusterNodes)
	options.AddJumpServerFlags(flagSet, &k.JumpServer)
	options.AddResetFlags(flagSet, &k.Reset)
}

func newResetOptions() *runOptions {
	kubeiOptions := options.NewKubei()
	kubeadmOptions := options.NewKubeadm()

	return &runOptions{
		kubei:   kubeiOptions,
		kubeadm: kubeadmOptions,
	}
}

func newResetData(cmd *cobra.Command, args []string, options *runOptions, out io.Writer) (*runData, error) {
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
