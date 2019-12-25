package cmd

import (
	"io"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	phases "github.com/yuyicai/kubei/cmd/phases/reset"
	"github.com/yuyicai/kubei/config/options"
	"github.com/yuyicai/kubei/config/rundata"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
)

// resetOptions defines all the init options exposed via flags by kubei reset.
type resetOptions struct {
	kubei   *options.Kubei
	kubeadm *options.Kubeadm
}

// compile-time assert that the local data object satisfies the phases data interface.
var _ phases.ResetData = &resetData{}

type resetData struct {
	kubei   *rundata.Kubei
	kubeadm *rundata.Kubeadm
}

// NewCmdreset returns "kubei reset" command.
func NewCmdreset(out io.Writer, resetOptions *resetOptions) *cobra.Command {
	if resetOptions == nil {
		resetOptions = newResetOptions()
	}
	resetRunner := workflow.NewRunner()

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Run this command in order to reset nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := resetRunner.InitData(args)
			if err != nil {
				return err
			}

			if err := resetRunner.Run(args); err != nil {
				return err
			}

			return nil
		},
		Args: cobra.NoArgs,
	}

	// adds flags to the reset command
	// reset command local flags could be eventually inherited by the sub-commands automatically generated for phases
	addResetConfigFlags(cmd.Flags(), resetOptions.kubei)
	options.AddKubeadmConfigFlags(cmd.Flags(), resetOptions.kubeadm)

	// initialize the workflow runner with the list of phases
	resetRunner.AppendPhase(phases.NewResetPhase())

	// sets the rundata builder function, that will be used by the runner
	// both when running the entire workflow or single phases
	resetRunner.SetDataInitializer(func(cmd *cobra.Command, args []string) (workflow.RunData, error) {
		return newResetData(cmd, args, resetOptions, out)
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
}

func newResetOptions() *resetOptions {
	kubeiOptions := options.NewKubei()
	kubeadmOptions := options.NewKubeadm()

	return &resetOptions{
		kubei:   kubeiOptions,
		kubeadm: kubeadmOptions,
	}
}

func newResetData(cmd *cobra.Command, args []string, options *resetOptions, out io.Writer) (*resetData, error) {

	kubeicfg := rundata.NewKubei()
	kubeadmcfg := rundata.NewKubeadm()

	options.kubei.ApplyTo(kubeicfg)
	options.kubeadm.ApplyTo(kubeadmcfg)

	resetDatacfg := &resetData{
		kubei:   kubeicfg,
		kubeadm: kubeadmcfg,
	}

	return resetDatacfg, nil
}

func (d *resetData) Cluster() *rundata.ClusterNodes {
	return &d.kubei.ClusterNodes
}

func (d *resetData) Cfg() *rundata.Kubei {
	return d.kubei
}

func (d *resetData) KubeadmCfg() *rundata.Kubeadm {
	return d.kubeadm
}
