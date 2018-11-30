package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/client"
)

var describeImageCmd = &cobra.Command{
	Use:   "describe-image",
	Short: "Get image info as a Json string",
	Long:  `Get a json representation of an image as it is stored in unik.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if host == "" {
				host = clientConfig.Host
			}
			if name == "" {
				return errors.New("must specify --image", nil)
			}
			image, err := client.UnikClient(host).Images().Get(name)
			if err != nil {
				return err
			}
			data, err := json.Marshal(image)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", string(data))
			return nil
		}(); err != nil {
			logrus.Errorf("describing image failed: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(describeImageCmd)
	describeImageCmd.Flags().StringVar(&name, "image", "", "<string,required> name or id of image. unik accepts a prefix of the name or id")
}
