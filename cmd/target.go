package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"errors"
	"io/ioutil"
	"github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
)

var show bool

var targetCmd = &cobra.Command{
	Use:   "target",
	Short: "Configure unik daemon URL for cli client commands",
	Long: `Sets the url of the unik daemon for cli commands.
	If running unik locally, say 'unik target -url localhost'

	args:
	--url: <string, required>: url/ip address of the host running the unik daemon
	--port: <int, optional>: port the daemon is running on (default: 3000)

	--show: <bool,optional>: shows the current target that is set`,
	Run: func(cmd *cobra.Command, args []string) {
		if show {
			readClientConfig()
			logrus.Infof("Current target: %s", clientConfig.DaemonUrl)
			return
		}
		if url == "" {
			logrus.Error("--url must be set for target")
			os.Exit(-1)
		}
		if err := setClientConfig(url, port); err != nil {
			logrus.WithError(err).Error("failed to save target to config file")
			os.Exit(-1)
		}
		logrus.Infof("target set: %s:%v", url, port)
	},
}

func setClientConfig(url string, port int) error {
	data, err := yaml.Marshal(config.ClientConfig{DaemonUrl: fmt.Sprintf("%s:%v", url, port)})
	if err != nil {
		return errors.New("failed to convert config to yaml string: "+err.Error())
	}
	if err := ioutil.WriteFile(clientConfigFile, data, 0644); err !=nil {
		return errors.New("failed writing config to file "+ clientConfigFile +": "+err.Error())
	}
	return nil
}

func init() {
	RootCmd.AddCommand(targetCmd)
	targetCmd.Flags().BoolVar(&show, "show", false, "<bool,optional>: shows the current target that is set")
}
