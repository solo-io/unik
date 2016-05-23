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
)

func DaemonFromEnv() (*daemon.UnikDaemon, error) {
	var daemonConfig config.DaemonConfig
	daemonConfigString := os.Getenv("DAEMON_CONFIG")
	if err := yaml.Unmarshal([]byte(daemonConfigString), &daemonConfig); err != nil {
		return nil, errors.New("not valid yaml: "+daemonConfigString, err)
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