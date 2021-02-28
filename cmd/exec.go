package cmd

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/yuyicai/kubei/internal/operator"
	"github.com/yuyicai/kubei/internal/options"
	"github.com/yuyicai/kubei/internal/preflight"
	"github.com/yuyicai/kubei/internal/rundata"
)

// NewCmdExec returns "kubei exec" command.
func NewCmdExec(out io.Writer, runOptions *runOptions) *cobra.Command {
	if runOptions == nil {
		runOptions = newExecOptions()
	}

	cluster := &rundata.Cluster{}

	var command string

	cmd := &cobra.Command{
		Use:   "exec",
		Short: "execute command on nodes",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			date, err := newExecData(runOptions)
			if err != nil {
				return err
			}
			cluster = date.Cluster()
			return preflight.ExecPrepare(cluster)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExec(cluster, command)
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			preflight.CloseSSH(cluster)
			return nil
		},
		Args: cobra.NoArgs,
	}

	// adds flags to the exec command
	// exec command local flags could be eventually inherited by the sub-commands automatically generated for phases
	addExecConfigFlags(cmd.Flags(), runOptions.kubei, &command)

	return cmd
}

func addExecConfigFlags(flagSet *flag.FlagSet, k *options.Kubei, command *string) {
	options.AddPublicUserInfoConfigFlags(flagSet, &k.ClusterNodes.PublicHostInfo)
	options.AddKubeClusterNodesConfigFlags(flagSet, &k.ClusterNodes)
	options.AddJumpServerFlags(flagSet, &k.JumpServer)
	options.AddExecCommandFlags(flagSet, command)
}

func newExecOptions() *runOptions {
	kubeiOptions := options.NewKubei()
	kubeadmOptions := options.NewKubeadm()

	return &runOptions{
		kubei:   kubeiOptions,
		kubeadm: kubeadmOptions,
	}
}

func newExecData(options *runOptions) (*runData, error) {
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

func runExec(c *rundata.Cluster, command string) error {
	if command == "" {
		return errors.New("the command is empty, please use the flag \"--command\" to set command")
	}
	fmt.Println(color.HiBlueString("Executing command:"), color.HiYellowString(command))
	return operator.RunOnAllNodes(c, func(node *rundata.Node, c *rundata.Cluster) error {
		if err := node.Run(command); err != nil {
			return errors.Wrapf(err, "[%s] [exec] Failed to execute command: %s", node.HostInfo.Host, command)
		}
		fmt.Println(fmt.Sprintf("[%s] [exec] execute command: %s", node.HostInfo.Host, color.HiGreenString("done✅️")))
		return nil
	})
}
