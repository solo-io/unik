package cmd

import (
	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"fmt"
	"errors"
)

var psCmd = &cobra.Command{
	Use:   "instances",
	Aliases: []string{"ps"},
	Short: "List pending/running/stopped unik instances",
	Long: `Lists all unik-managed instances across providers.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if host == "" {
				host = clientConfig.Host
			}
			logrus.WithField("host", host).Info("listing images")
			instances, err := client.UnikClient(host).Instances().All()
			if err != nil {
				return errors.New(fmt.Sprintf("listing images failed: %v", err))
			}
			printInstances(instances...)
			return nil
		}(); err != nil {
			logrus.Errorf("failed listing instances: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(psCmd)
}
