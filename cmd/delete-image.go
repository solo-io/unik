package cmd

import (
	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"os"
	"github.com/emc-advanced-dev/pkg/errors"
)

var rmiCmd = &cobra.Command{
	Use:   "delete-image",
	Aliases: []string{"rmi"},
	Short: "Delete a unikernel image",
	Long: `Deletes an image.
	You may specify the image by name or id.`,
	
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if imageName == "" {
				return errors.New("must specify --image", nil)
			}
			if host == "" {
				host = clientConfig.Host
			}
			logrus.WithFields(logrus.Fields{"host": host, "force": force, "image": imageName}).Info("deleting image")
			if err := client.UnikClient(host).Images().Delete(imageName, force); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			logrus.Errorf("failed deleting image: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(rmiCmd)
	rmiCmd.Flags().StringVar(&imageName, "image", "", "<string,required> name or id of image. unik accepts a prefix of the name or id")
	rmiCmd.Flags().BoolVar(&force, "force", false, "<bool, optional> force deleting image in the case that it is running")
}
