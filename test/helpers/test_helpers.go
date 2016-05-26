package helpers

import (
	"github.com/emc-advanced-dev/unik/pkg/daemon"
	"gopkg.in/yaml.v2"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/pkg/errors"
	"path/filepath"
	"os/exec"
	"github.com/emc-advanced-dev/unik/pkg/util"
	"io/ioutil"
)

func DaemonFromEnv() (*daemon.UnikDaemon, error) {
	var daemonConfig config.DaemonConfig
	var data []byte
	daemonConfigFile := os.Getenv("DAEMON_CONFIG_FILE")
	if daemonConfigFile == "" {
		daemonConfigFile = os.Getenv("HOME")+"/.unik/daemon-config.yaml"
	}
	data, err := ioutil.ReadFile(daemonConfigFile)
	if err != nil {
		return nil, errors.New("failed to read "+daemonConfigFile, err)
	}
	if err := yaml.Unmarshal(data, &daemonConfig); err != nil {
		return nil, errors.New("not valid yaml: "+ daemonConfigFile, err)
	}
	d, err := daemon.NewUnikDaemon(daemonConfig)
	if err != nil {
		return nil, errors.New("daemon failed to initialize", err)
	}
	return d, nil
}

func KillUnikstate() error {
	return os.RemoveAll(filepath.Join(os.Getenv("HOME"), ".unik"))
}

func MakeContainers(projectRoot string) error {
	cmd := exec.Command("make", "-C", projectRoot, "containers")
	util.LogCommand(cmd, false)
	return cmd.Run()
}

func RemoveContainers(projectRoot string) error {
	cmd := exec.Command("make", "-C", projectRoot, "remove-containers")
	util.LogCommand(cmd, false)
	return cmd.Run()
}