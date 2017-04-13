package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/client"
)

var describeInstanceCmd = &cobra.Command{
	Use:   "describe-instance",
	Short: "Get instance info as a Json string",
	Long:  `Get a json representation of an instance as it is stored in unik.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if host == "" {
				host = clientConfig.Host
			}
			if name == "" {
				return errors.New("must specify --instance", nil)
			}
			instance, err := client.UnikClient(host).Instances().Get(name)
			if err != nil {
				return err
			}
			data, err := json.Marshal(instance)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", string(data))
			return nil
		}(); err != nil {
			logrus.Errorf("failed describing instance: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(describeInstanceCmd)
	describeInstanceCmd.Flags().StringVar(&name, "instance", "", "<string,required> name or id of instance. unik accepts a prefix of the name or id")
}
