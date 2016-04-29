package cmd

import (
	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/client"
)

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "instances",
	Aliases: []string{"ps"},
	Short: "List pending/running/stopped unik instances",
	Long: `Lists all unik-managed instances across providers.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		readClientConfig()
		if host == "" {
			host = clientConfig.Host
		}
		logrus.WithField("host", host).Info("listing images")
		instances, err := client.UnikClient(host).Instances().All()
		if err != nil {
			logrus.WithError(err).Error("listing images failed")
			os.Exit(-1)
		}
		printInstances(instances...)
	},
}

func init() {
	RootCmd.AddCommand(psCmd)
}
