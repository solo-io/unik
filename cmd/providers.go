package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/client"
)

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List available unikernel providers",
	Long:  `Returns a list of providers available to the targeted unik backend.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if host == "" {
				host = clientConfig.Host
			}
			logrus.WithField("host", host).Info("listing providers")
			providers, err := client.UnikClient(host).AvailableProviders()
			if err != nil {
				return errors.New("listing providers failed", err)
			}
			fmt.Printf("%s\n", strings.Join(providers, "\n"))
			return nil
		}(); err != nil {
			logrus.Errorf("failed listing providers: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(providersCmd)
}
