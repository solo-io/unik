package cmd

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/client"
)

var detachCmd = &cobra.Command{
	Use:     "detach-volume",
	Aliases: []string{"detach"},
	Short:   "Detach an attached volume from a stopped instance",
	Long: `Detaches a volume to a stopped instance at a specified mount point.
You specify the volume by name or id.

After detaching the volume, the volume can be mounted to another instance.

If the instance is not stopped, detach will result in an error.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if volumeName == "" {
				return errors.New("must specify --volume", nil)
			}
			if host == "" {
				host = clientConfig.Host
			}
			logrus.WithFields(logrus.Fields{"host": host, "volume": volumeName}).Info("detaching volume")
			if err := client.UnikClient(host).Volumes().Detach(volumeName); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			logrus.Errorf("failed deleting volume: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(detachCmd)
	detachCmd.Flags().StringVar(&volumeName, "volume", "", "<string,required> name or id of volume to detach. unik accepts a prefix of the name or id")
}
