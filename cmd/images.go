package cmd

import (
	"github.com/spf13/cobra"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"os"
	"github.com/Sirupsen/logrus"
)

// imagesCmd represents the images command
var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "List available unikernel images",
	Long: `Lists all available unikernel images across providers.
	Includes important information for running and managing instances,
	including bind mounts required at runtime.`,
	Run: func(cmd *cobra.Command, args []string) {
		readClientConfig()
		if url == "" {
			url = clientConfig.DaemonUrl
		}
		logrus.WithField("url", url).Info("listing images")
		images, err := client.UnikClient(url).Images().All()
		if err != nil {
			logrus.WithError(err).Error("listing images failed")
			os.Exit(-1)
		}
		printImages(images...)
	},
}

func init() {
	RootCmd.AddCommand(imagesCmd)
}
