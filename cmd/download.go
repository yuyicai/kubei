package cmd

import (
	"github.com/yuyicai/kubei/internal/phases/download"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/klog"
)

const DefaultKubernetesVersion = "v1.20.0"

func NewCmdDownload(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "download kubernetes files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDownload(out, cmd)
		},
	}
	cmd.Flags().StringP("kube-version", "", "v1.20.0", "kubernetes version")
	return cmd
}

func RunDownload(out io.Writer, cmd *cobra.Command) error {
	klog.V(1).Infoln("download kubernetes files")
	return download.KubeFiles(DefaultKubernetesVersion, "")
}
