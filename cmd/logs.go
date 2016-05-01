package cmd

import (
	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"os"
	"errors"
	"fmt"
	"bufio"
)

var follow, deleteOnDisconnect bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "retrieve the logs (stdout) of a unikernel instance",
	Long: `Retrieves logs from a running unikernel instance.

	Cannot be used on an instance in powered-off state.
	Use the --follow flag to attach to the instance's stdout
	Use --delete in combination with --follow to force automatic instance
	deletion when the HTTP connection to the instance is broken (by client
	disconnect). The --delete flag is typically intended for use with
	orchestration software such as cluster managers which may require
	a persistent http connection managed instances.

	You may specify the instance by name or id.

	Example usage:
		unik logs --instancce myInstance

		# will return captured stdout from myInstance since boot time

		unik logs --instance myInstance --follow --delete

		# will open an http connection between the cli and unik
		backend which streams stdout from the instance to the client
		# when the client disconnects (i.e. with Ctrl+C) unik will
		automatically power down and terminate the instance
		`,
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
				r, err := client.UnikClient(host).Instances().AttachLogs(instanceName, deleteOnDisconnect)
				if err != nil {
					return err
				}
				reader := bufio.NewReader(r)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						return err
					}
					fmt.Printf(line)
				}
			} else {
				logrus.WithFields(logrus.Fields{"host": host, "instance": instanceName}).Info("getting instance logs")
				data, err := client.UnikClient(host).Instances().GetLogs(instanceName)
				if err != nil {
					return err
				}
				fmt.Printf("%s\n", string(data))
			}
			return nil
		}(); err != nil {
			logrus.Errorf("failed retrieving instance logs: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringVar(&instanceName, "instance", "", "<string,required> name or id of instance. unik accepts a prefix of the name or id")
	logsCmd.Flags().BoolVar(&follow, "follow", false, "<bool,optional> follow stdout of instance as it is printed")
	logsCmd.Flags().BoolVar(&deleteOnDisconnect, "delete", false, "<bool,optional> use this flag with the --follow flag to trigger automatic deletion of instance after client closes the http connection")
}
