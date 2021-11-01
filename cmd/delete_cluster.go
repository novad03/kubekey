package cmd

import (
	common2 "github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/spf13/cobra"
)

var deleteClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Delete a cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := common2.Argument{
			FilePath: opt.ClusterCfgFile,
			Debug:    opt.Verbose,
		}
		return pipelines.DeleteCluster(arg)
	},
}

func init() {
	deleteCmd.AddCommand(deleteClusterCmd)

	deleteClusterCmd.Flags().StringVarP(&opt.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
