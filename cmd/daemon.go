package cmd

import (
	uniklog "github.com/emc-advanced-dev/unik/pkg/util/log"
	"github.com/spf13/cobra"
	"os"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"gopkg.in/yaml.v2"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
)

var daemonConfigFile, logFile string
var debugMode, trace bool

// daemonCmd represents the daemon command
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

	flags:
		--daemon-config: <string, required> configuration file for daemon (default is $HOME/.unik/daemon-config.yaml)
		--port: <int, optional> port for daemon to run on. 3000 by default
		--debug: <bool, optional> more verbose logging for the daemon
		--trace: <bool, optional> add stack trace to daemon logs
		--logfile: <string, optional> output logs to file (in addition to stdout)

	Example usage:
		unik daemon -daemon-config ./my-config.yaml -port 12345 -debug -trace -logfile logs.txt

		 # will start the daemon using config file at my-config.yaml
		 # running on port 12345
		 # debug mode activated
		 # trace mode activated
		 # outputting logs to logs.txt
`,
	Run: func(cmd *cobra.Command, args []string) {
		readDaemonConfig()
		if debugMode {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if trace {
			logrus.AddHook(&uniklog.AddTraceHook{true})
		}
		if logFile != "" {
			f, err := os.Open(logFile)
			if err != nil {
				logrus.WithError(err).Errorf("failed to open log file %s for writing", logFile)
				os.Exit(-1)
			}
			logrus.AddHook(&uniklog.TeeHook{f})
		}
		logrus.WithField("config", daemonConfig).Info("daemon started")
		d, err := daemon.NewUnikDaemon(daemonConfig)
		if err != nil {
			logrus.WithError(err).Errorf("daemon failed to initialize")
			os.Exit(-1)
		}
		d.Run(port)
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().StringVar(&daemonConfigFile, "daemon-config", os.Getenv("HOME")+"/.unik/daemon-config.yaml", "daemon config file (default is $HOME/.unik/daemon-config.yaml)")
	daemonCmd.Flags().IntVar(&port, "port", 3000, "<int, optional> listening port for daemon")
	daemonCmd.Flags().BoolVar(&debugMode, "debug", false, "<bool, optional> more verbose logging for the daemon")
	daemonCmd.Flags().BoolVar(&trace, "trace", false, "<bool, optional> add stack trace to daemon logs")
	daemonCmd.Flags().StringVar(&logFile, "logfile", "", "<string, optional> output logs to file (in addition to stdout)")
}


var daemonConfig config.DaemonConfig
func readDaemonConfig() {
	data, err := ioutil.ReadFile(daemonConfigFile)
	if err != nil {
		logrus.WithError(err).Errorf("failed to read daemon configuration file at "+ daemonConfigFile +`\n
		See documentation at http://github.com/emc-advanced-dev/unik for creating daemon config.'`)
		os.Exit(-1)
	}
	if err := yaml.Unmarshal(data, &daemonConfig); err != nil {
		logrus.WithError(err).Errorf("failed to parse daemon configuration yaml at "+ daemonConfigFile +`\n
		Please ensure config file contains valid yaml.'`)
		os.Exit(-1)
	}
}
