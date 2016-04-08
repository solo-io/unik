package main

import (
	"flag"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/layer-x/layerx-commons/lxlog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
)

func main() {
	debugMode := flag.String("debug", "false", "enable verbose/debug mode")
	port := flag.Int("port", 3000, "port to run unik daemon on")
	flag.Parse()
	if *debugMode == "true" {
		lxlog.GlobalLogLevel = lxlog.DebugLevel
	}
	logger := lxlog.New("unik-daemon-main")

	buildCommand := exec.Command("make")
	buildCommand.Dir = "../../containers/"
	logger.LogCommand(buildCommand, true)
	err := buildCommand.Run()
	if err != nil {
		logger.WithErr(err).Fatalf("building containers")
	}

	logger.Infof("all images finished")

	configData, err := ioutil.ReadFile("conf.yml")
	if err != nil {
		logger.WithErr(err).Fatalf("reading config file conf.yml")
	}

	var config config.UnikConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		logger.WithErr(err).Fatalf("parsing conf.yml")
	}

	unikDaemon := daemon.NewUnikDaemon(config)
	unikDaemon.Run(logger, *port)
}
