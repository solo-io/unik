package cmd

import (
	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"os"
	"errors"
	"fmt"
)

var follow, deleteOnDisconnect bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "logs a running unikernel instance",
	Long: `logss a running instance.
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
				return errors.New("must specify --instance")
			}
			if follow {
				logrus.WithFields(logrus.Fields{"host": host, "instance": instanceName}).Info("attaching to instance")
				w, err := client.UnikClient(host).Instances().AttachLogs(instanceName, deleteOnDisconnect)
				if err != nil {
					return err
				}

			} else {

			}
			logrus.WithFields(logrus.Fields{"host": host, "instance": instanceName}).Info("getting instance logs")
			data, err := client.UnikClient(host).Instances().GetLogs(instanceName)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", string(data))
			return nil
		}(); err != nil {
			logrus.Errorf("failed logsping instance: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringVar(&instanceName, "instance", "", "<string,required> name or id of instance. unik accepts a prefix of the name or id")
	logsCmd.Flags().BoolVar(&follow, "instance", false, "<bool,optional> follow stdout of instance as it is printed")
	logsCmd.Flags().BoolVar(&deleteOnDisconnect, "delete", false, "<bool,optional> use this flag with the --follow flag to trigger automatic deletion of instance after client closes the http connection")
}
