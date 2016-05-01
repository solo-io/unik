package cmd

import (
	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"os"
	"errors"
)

var volumeName string 

var rmvCmd = &cobra.Command{
	Use:   "delete-volume",
	Aliases: []string{"rmv"},
	Short: "Delete a unikernel volume",
	Long: `Deletes an volume.
	You may specify the volume by name or id.`,
	
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if volumeName == "" {
				return errors.New("must specify --volume")
			}
			if host == "" {
				host = clientConfig.Host
			}
			logrus.WithFields(logrus.Fields{"host": host, "force": force, "volume": volumeName}).Info("deleting volume")
			if err := client.UnikClient(host).Volumes().Delete(volumeName, force); err != nil {
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
	RootCmd.AddCommand(rmvCmd)
	rmvCmd.Flags().StringVar(&volumeName, "volume", "", "<string,required> name or id of volume. unik accepts a prefix of the name or id")
	rmvCmd.Flags().BoolVar(&force, "force", false, "<bool, optional> force deleting volume in the case that it is running")
}
