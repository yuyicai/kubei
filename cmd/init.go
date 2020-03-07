package cmd

import (
	"io"

	phases "github.com/yuyicai/kubei/cmd/phases/init"
	"github.com/yuyicai/kubei/config/options"
	"github.com/yuyicai/kubei/config/rundata"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
)

// initOptions defines all the init options exposed via flags by kubei init.
type initOptions struct {
	kubei   *options.Kubei
	kubeadm *options.Kubeadm
}

// compile-time assert that the local data object satisfies the phases data interface.
var _ phases.InitData = &initData{}

// initData defines all the runtime information used when running the kubei init workflow;
// this data is shared across all the phases that are included in the workflow.
type initData struct {
	kubei   *rundata.Kubei
	kubeadm *rundata.Kubeadm
}

// NewCmdInit returns "kubei init" command.
func NewCmdInit(out io.Writer, initOptions *initOptions) *cobra.Command {
	if initOptions == nil {
		initOptions = newInitOptions()
	}
	initRunner := workflow.NewRunner()

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Run this command in order to create a high availability Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := initRunner.InitData(args)
			if err != nil {
				return err
			}

			if err := initRunner.Run(args); err != nil {
				return err
			}

			return nil
		},
		Args: cobra.NoArgs,
	}

	// adds flags to the init command
	// init command local flags could be eventually inherited by the sub-commands automatically generated for phases
	addInitConfigFlags(cmd.Flags(), initOptions.kubei)
	options.AddKubeadmConfigFlags(cmd.Flags(), initOptions.kubeadm)

	// initialize the workflow runner with the list of phases
	initRunner.AppendPhase(phases.NewContainerEnginePhase())
	initRunner.AppendPhase(phases.NewKubeComponentPhase())
	initRunner.AppendPhase(phases.NewKubeadmPhase())

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
}

func newInitOptions() *initOptions {
	kubeiOptions := options.NewKubei()
	kubeadmOptions := options.NewKubeadm()

	return &initOptions{
		kubei:   kubeiOptions,
		kubeadm: kubeadmOptions,
	}
}

func newInitData(cmd *cobra.Command, args []string, options *initOptions, out io.Writer) (*initData, error) {

	kubeicfg := rundata.NewKubei()
	kubeadmcfg := rundata.NewKubeadm()

	options.kubei.ApplyTo(kubeicfg)
	options.kubeadm.ApplyTo(kubeadmcfg)

	rundata.DefaulKubeiConf(kubeicfg)
	rundata.DefaulkubeadmConf(kubeadmcfg)

	initDatacfg := &initData{
		kubei:   kubeicfg,
		kubeadm: kubeadmcfg,
	}

	return initDatacfg, nil
}

func (d *initData) Cluster() *rundata.ClusterNodes {
	return &d.kubei.ClusterNodes
}

func (d *initData) ContainerEngine() *rundata.ContainerEngine {
	return &d.kubei.ContainerEngine
}

func (d *initData) Cfg() *rundata.Kubei {
	return d.kubei
}

func (d *initData) KubeadmCfg() *rundata.Kubeadm {
	return d.kubeadm
}
