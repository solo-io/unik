package cmd

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push an image to a Unik Image Repository",
	Long: `
Example usage:
unik push --image myImage

Requires that you first authenticate to a unik image repository with 'unik login'
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := readClientConfig(); err != nil {
			logrus.Fatal(err)
		}
		c, err := getHubConfig()
		if err != nil {
			logrus.Fatal(err)
		}
		if imageName == "" {
			logrus.Fatal("--image must be set")
		}
		if host == "" {
			host = clientConfig.Host
		}
		if err := client.UnikClient(host).Images().Push(c, imageName); err != nil {
			logrus.Fatal(err)
		}
		fmt.Println(imageName + " pushed")
	},
}

func getHubConfig() (config.HubConfig, error) {
	var c config.HubConfig
	data, err := ioutil.ReadFile(hubConfigFile)
	if err != nil {
		return c, errors.New("reading "+hubConfigFile, err)
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return c, errors.New("failed to convert config from yaml", err)
	}
	return c, nil
}

func init() {
	RootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVar(&imageName, "image", "", "<string,required> image to push")
}
