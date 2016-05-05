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
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/unik/pkg/client"
)

var compilersCmd = &cobra.Command{
	Use:   "compilers",
	Short: "List available unikernel compilers",
	Long:  `Returns a list of compilers available to the targeted unik backend.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if host == "" {
				host = clientConfig.Host
			}
			logrus.WithField("host", host).Info("listing compilers")
			compilers, err := client.UnikClient(host).AvailableCompilers()
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", strings.Join(compilers, "\n"))
			return nil
		}(); err != nil {
			logrus.Errorf("failed listing compilers: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(compilersCmd)
}
