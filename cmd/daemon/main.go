package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
	uniklog "github.com/emc-advanced-dev/unik/pkg/util/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
)

func main() {
	debugMode := flag.Bool("debug", false, "enable verbose/debug mode")
	stackTrace := flag.Bool("trace", false, "additional debug option to add full stack trace to logs")
	logFile := flag.String("log", "", "optional file to write logs to")
	port := flag.Int("port", 3000, "port to run unik daemon on")
	flag.Parse()
	if *debugMode {
		logrus.SetLevel(logrus.DebugLevel)
		if *stackTrace {
			logrus.AddHook(&uniklog.AddTraceHook{true})
		} else {
			logrus.AddHook(&uniklog.AddTraceHook{false})
		}
	}
	if *logFile != "" {
		f, err := os.Open(*logFile)
		if err != nil {
			logrus.WithError(err).Fatalf("failed to open log file for writing")
		}
		logrus.AddHook(&uniklog.TeeHook{f})
	}

	buildCommand := exec.Command("make")
	buildCommand.Dir = "../../containers/"
	uniklog.LogCommand(buildCommand, true)
	if err := buildCommand.Run(); err != nil {
		logrus.WithError(err).Fatalf("building containers")
	}

	logrus.Infof("all images finished")

	configData, err := ioutil.ReadFile("conf.yml")
	if err != nil {
		logrus.WithError(err).Fatalf("reading config file conf.yml")
	}

	var config config.UnikConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		logrus.WithError(err).Fatalf("parsing conf.yml")
	}

	unikDaemon := daemon.NewUnikDaemon(config)
	unikDaemon.Run(*port)
}
