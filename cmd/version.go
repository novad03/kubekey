/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"github.com/kubesphere/kubekey/version"
	"github.com/spf13/cobra"
	"io"
)

var shortVersion bool
var showSupportedK8sVersionList bool

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the client version information",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if showSupportedK8sVersionList {
			return printSupportedK8sVersionList(cmd.OutOrStdout())
		}
		return printVersion(shortVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&shortVersion, "short", "", false, "print the version number")
	versionCmd.Flags().BoolVarP(&showSupportedK8sVersionList, "show-supported-k8s", "", false,
		`print the version of supported k8s`)
}

func printVersion(short bool) error {
	v := version.Get()
	if short {
		if len(v.GitCommit) >= 7 {
			fmt.Printf("%s+g%s\n", v.Version, v.GitCommit[:7])
			return nil
		}
		fmt.Println(version.GetVersion())
	}
	fmt.Printf("%#v\n", v)
	return nil
}

func printSupportedK8sVersionList(output io.Writer) (err error) {
	_, err = output.Write([]byte(`v1.15.12
v1.16.8
v1.16.10
v1.16.12
v1.16.13
v1.17.0
v1.17.4
v1.17.5
v1.17.6
v1.17.7
v1.17.8
v1.17.9
v1.18.3
v1.18.5
v1.18.6
v1.18.8
v1.19.0
`))
	return
}
