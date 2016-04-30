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
	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"os"
	"errors"
)

// rmiCmd represents the rmi command
var rmiCmd = &cobra.Command{
	Use:   "delete-image",
	Aliases: []string{"rmi"},
	Short: "Delete a unikernel image",
	Long: `Deletes an image.
	You may specify the image by name or id.`,
	
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if imageName == "" {
				return errors.New("must specify --name")
			}
			if host == "" {
				host = clientConfig.Host
			}
			logrus.WithFields(logrus.Fields{"host": host, "force": force, "image": imageName}).Info("deleting image")
			if err := client.UnikClient(host).Images().Delete(imageName, force); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			logrus.Errorf("failed deleting image: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(rmiCmd)
	rmiCmd.Flags().StringVar(&imageName, "image", "", "<string,required> name or id of image. unik accepts a prefix of the name or id")
	rmiCmd.Flags().BoolVar(&force, "force", false, "<bool, optional> force deleting image in the case that it is running")
}
