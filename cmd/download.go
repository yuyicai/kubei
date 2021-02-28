package cmd

import (
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/internal/phases/download"
)

const DefaultKubernetesVersion = "v1.20.4"

func NewCmdDownload(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "download kubernetes files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDownload(out, cmd)
		},
	}
	cmd.Flags().String("kube-version", DefaultKubernetesVersion, "kubernetes version")
	return cmd
}

func RunDownload(out io.Writer, cmd *cobra.Command) error {
	klog.V(1).Infoln("download kubernetes files")
	version, err := cmd.Flags().GetString("kube-version")
	if err != nil {
		return errors.Wrapf(err, "error accessing flag %s for command %s", "kube-version", cmd.Name())
	}
	return download.KubeFiles(version, "")
}
