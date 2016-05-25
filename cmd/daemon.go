package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"net/url"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
)

var daemonRuntimeFolder, daemonConfigFile, logFile string
var debugMode, trace bool

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Runs the unik backend (daemon)",
	Long: `Run this command to start the unik daemon process.
This should normally be left running as a long-running background process.
The daemon requires that docker is installed and running on the your system.
Necessary docker containers must be built for the daemon to work properly;
Run 'make' in the unik root directory to build all necessary containers.

Daemon also requires a configuration file with credentials and configuration info
for desired providers.

Example usage:
	unik daemon --f ./my-config.yaml --port 12345 --debug --trace --logfile logs.txt

	 # will start the daemon using config file at my-config.yaml
	 # running on port 12345
	 # debug mode activated
	 # trace mode activated
	 # outputting logs to logs.txt
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {

			// set unik home
			config.Internal.UnikHome = daemonRuntimeFolder

			if daemonConfigFile == "" {
				daemonConfigFile = filepath.Join(config.Internal.UnikHome, "daemon-config.yaml")
			}

			if err := readDaemonConfig(); err != nil {
				return err
			}

			//don't print vsphere password
			redactions := []string{}
			for _, vsphereConfig := range daemonConfig.Providers.Vsphere {
				redactions = append(redactions, vsphereConfig.VspherePassword, url.QueryEscape(vsphereConfig.VspherePassword))
			}
			logrus.SetFormatter(&unikutil.RedactedTextFormatter{
				Redactions: redactions,
			})

			if debugMode {
				logrus.SetLevel(logrus.DebugLevel)
			}
			if trace {
				logrus.AddHook(&unikutil.AddTraceHook{true})
			}
			if logFile != "" {
				os.Create(logFile)
				f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
				if err != nil {
					return errors.New(fmt.Sprintf("failed to open log file %s for writing", logFile), err)
				}
				logrus.AddHook(&unikutil.TeeHook{f})
			}

			logrus.WithField("config", daemonConfig).Info("daemon started")
			d, err := daemon.NewUnikDaemon(daemonConfig)
			if err != nil {
				return errors.New("daemon failed to initialize", err)
			}
			d.Run(port)
			return nil
		}(); err != nil {
			logrus.Errorf("running daemon failed: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().StringVar(&daemonRuntimeFolder, "d", os.Getenv("HOME")+"/.unik/", "daemon runtime folder - where state is stored. (default is $HOME/.unik/)")
	daemonCmd.Flags().StringVar(&daemonConfigFile, "f", "", "daemon config file (default is {RuntimeFolder}/daemon-config.yaml)")
	daemonCmd.Flags().IntVar(&port, "port", 3000, "<int, optional> listening port for daemon")
	daemonCmd.Flags().BoolVar(&debugMode, "debug", false, "<bool, optional> more verbose logging for the daemon")
	daemonCmd.Flags().BoolVar(&trace, "trace", false, "<bool, optional> add stack trace to daemon logs")
	daemonCmd.Flags().StringVar(&logFile, "logfile", "", "<string, optional> output logs to file (in addition to stdout)")
}

var daemonConfig config.DaemonConfig

func readDaemonConfig() error {
	data, err := ioutil.ReadFile(daemonConfigFile)
	if err != nil {
		logrus.WithError(err).Errorf("failed to read daemon configuration file at " + daemonConfigFile + `\n
		See documentation at http://github.com/emc-advanced-dev/unik for creating daemon config.'`)
		return err
	}
	if err := yaml.Unmarshal(data, &daemonConfig); err != nil {
		logrus.WithError(err).Errorf("failed to parse daemon configuration yaml at " + daemonConfigFile + `\n
		Please ensure config file contains valid yaml.'`)
		return err
	}
	return nil
}
