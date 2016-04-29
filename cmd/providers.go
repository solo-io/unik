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

	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"strings"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/client"
)

// providersCmd represents the providers command
var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List available unikernel providers",
	Long: `Returns a list of providers available to the targeted unik backend.`,
	Run: func(cmd *cobra.Command, args []string) {
		readClientConfig()
		if host == "" {
			host = clientConfig.Host
		}
		logrus.WithField("host", host).Info("listing providers")
		providers, err := client.UnikClient(host).AvailableProviders()
		if err != nil {
			logrus.WithError(err).Error("listing providers failed")
			os.Exit(-1)
		}
		fmt.Printf("%s\n", strings.Join(providers, "\n"))
	},
}

func init() {
	RootCmd.AddCommand(providersCmd)
}
