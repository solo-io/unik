package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/client"
)

var rmCmd = &cobra.Command{
	Use:     "delete-instance",
	Aliases: []string{"rm"},
	Short:   "Delete a unikernel instance",
	Long: `Deletes an instance.
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
			logrus.WithFields(logrus.Fields{"host": host, "force": force, "instance": instanceName}).Info("deleting instance")
			if err := client.UnikClient(host).Instances().Delete(instanceName, force); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			logrus.Errorf("failed deleting instance: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(rmCmd)
	rmCmd.Flags().StringVar(&instanceName, "instance", "", "<string,required> name or id of instance. unik accepts a prefix of the name or id")
	rmCmd.Flags().BoolVar(&force, "force", false, "<bool, optional> force deleting instance in the case that it is running")
}
