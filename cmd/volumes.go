package cmd

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/client"
)

var volumesCmd = &cobra.Command{
	Use:   "volumes",
	Short: "List available unik-managed volumes",
	Long: `Lists all available unik-managed volumes across providers.

ATTACHED-INSTANCE gives the instance ID of the instance a volume
is attached to, if any. Only volumes that have no attachment are
available to be attached to an instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if host == "" {
				host = clientConfig.Host
			}
			logrus.WithField("host", host).Info("listing volumes")
			volumes, err := client.UnikClient(host).Volumes().All()
			if err != nil {
				return errors.New("listing volumes failed", err)
			}
			printVolumes(volumes...)
			return nil
		}(); err != nil {
			logrus.Errorf("failed listing volumes: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(volumesCmd)
}
