package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"os"
	"github.com/Sirupsen/logrus"
	"encoding/json"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"github.com/emc-advanced-dev/pkg/errors"
)

var describeInstanceCmd = &cobra.Command{
	Use:   "describe-instance",
	Short: "Get instance info as a Json string",
	Long: `Get a json representation of an instance as it is stored in unik.`,
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
