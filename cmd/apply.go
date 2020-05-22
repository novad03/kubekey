package cmd

import (
	"github.com/kubesphere/kubekey/pkg/install"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/spf13/cobra"
)

func NewCmdScaleCluster() *cobra.Command {
	var (
		clusterCfgFile string
		//pkgDir         string
		verbose bool
	)
	var clusterCmd = &cobra.Command{
		Use:   "apply",
		Short: "Apply a configuration to a cluster by filename",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := util.InitLogger(verbose)
			//return scale.ScaleCluster(clusterCfgFile, logger, pkgDir, Verbose)
			return install.CreateCluster(clusterCfgFile, logger, false, verbose)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "file", "f", "", "configuration file name")
	//clusterCmd.Flags().StringVarP(&pkgDir, "pkg", "", "", "release package (offline)")
	clusterCmd.Flags().BoolVarP(&verbose, "debug", "", true, "")
	return clusterCmd
}
