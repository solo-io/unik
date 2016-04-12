package main

import (
	"flag"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
	"github.com/Sirupsen/logrus"
	uniklog "github.com/emc-advanced-dev/unik/pkg/util/log"
	"os"
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
	err := buildCommand.Run()
	if err != nil {
		logrus.WithError(err).Fatalf("building containers")
	}

	logrus.Infof("all images finished")

	configData, err := ioutil.ReadFile("conf.yml")
	if err != nil {
		logrus.WithError(err).Fatalf("reading config file conf.yml")
	}

	var config config.UnikConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		logrus.WithError(err).Fatalf("parsing conf.yml")
	}

	unikDaemon := daemon.NewUnikDaemon(config)
	unikDaemon.Run(*port)
}
