// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"github.com/spf13/cobra"
)

// remote-deleteCmd represents the remote-delete command
var remoteDeleteCmd = &cobra.Command{
	Use:   "remote-delete",
	Short: "Deleted a pushed image from a Unik Hub Repository",
	Long: `
Example usage:
unik remote-delete --image myImage

Requires that you first authenticate to a unik image repository with 'unik login'
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := readClientConfig(); err != nil {
			logrus.Fatal(err)
		}
		c, err := getHubConfig()
		if err != nil {
			logrus.Fatal(err)
		}
		if imageName == "" {
			logrus.Fatal("--imageName must be set")
		}
		if host == "" {
			host = clientConfig.Host
		}
		if err := client.UnikClient(host).Images().RemoteDelete(c, imageName); err != nil {
			logrus.Fatal(err)
		}
		fmt.Println(imageName + " pushed")
	},
}

func init() {
	RootCmd.AddCommand(remoteDeleteCmd)
	remoteDeleteCmd.Flags().StringVar(&imageName, "image", "", "<string,required> image to push")
}
