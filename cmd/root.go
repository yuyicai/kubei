package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
)

func NewKubeiCommand(in io.Reader, out, err io.Writer) *cobra.Command {

	cmds := &cobra.Command{
		Use:   "kubei",
		Short: "kubei: easily deploy a high availability Kubernetes cluster",
		Long:  "easily deploy a high availability Kubernetes cluster",
	}

	cmds.AddCommand(NewCmdInit(out, nil))
	cmds.AddCommand(NewCmdReset(out, nil))
	cmds.AddCommand(NewCmdVersion(out))
	return cmds

}

// Execute called by main.main().
func Execute() {

	klog.InitFlags(nil)
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Set("logtostderr", "true")
	// We do not want these flags to show up in --help
	// These MarkHidden calls must be after the lines above
	pflag.CommandLine.MarkHidden("version")
	pflag.CommandLine.MarkHidden("log-flush-frequency")
	pflag.CommandLine.MarkHidden("alsologtostderr")
	pflag.CommandLine.MarkHidden("log-backtrace-at")
	pflag.CommandLine.MarkHidden("log-dir")
	pflag.CommandLine.MarkHidden("logtostderr")
	pflag.CommandLine.MarkHidden("stderrthreshold")
	pflag.CommandLine.MarkHidden("vmodule")
	pflag.CommandLine.MarkHidden("add-dir-header")
	pflag.CommandLine.MarkHidden("log-file")
	pflag.CommandLine.MarkHidden("log-file-max-size")
	pflag.CommandLine.MarkHidden("skip-headers")
	pflag.CommandLine.MarkHidden("skip-log-headers")

	cmd := NewKubeiCommand(os.Stdin, os.Stdout, os.Stderr)

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
