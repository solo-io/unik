package cmd

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/client"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a stopped unikernel instance",
	Long: `Starts a stopped instance.
You may specify the instance by name or id.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if host == "" {
				host = clientConfig.Host
			}
			if instanceName == "" {
				return errors.New("must specify --instance", nil)
			}
			logrus.WithFields(logrus.Fields{"host": host, "instance": instanceName}).Info("starting instance")
			if err := client.UnikClient(host).Instances().Start(instanceName); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			logrus.Errorf("failed starting instance: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVar(&instanceName, "instance", "", "<string,required> name or id of instance. unik accepts a prefix of the name or id")
}
