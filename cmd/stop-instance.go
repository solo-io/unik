package cmd

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/client"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running unikernel instance",
	Long: `Stops a running instance.
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
			logrus.WithFields(logrus.Fields{"host": host, "instance": instanceName}).Info("stopping instance")
			if err := client.UnikClient(host).Instances().Stop(instanceName); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			logrus.Errorf("failed stopping instance: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(stopCmd)
	stopCmd.Flags().StringVar(&instanceName, "instance", "", "<string,required> name or id of instance. unik accepts a prefix of the name or id")
}
