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
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"github.com/Sirupsen/logrus"
	"runtime"
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
	util.LogCommand(cmd, true)
	return cmd.Run()
}

func RemoveContainers(projectRoot string) error {
	cmd := exec.Command("make", "-C", projectRoot, "remove-containers")
	util.LogCommand(cmd, false)
	return cmd.Run()
}

func TarExampleApp(projectRoot string, appDir string) (*os.File, error) {
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, errors.New("getting abs of "+projectRoot, err)
	}
	path := filepath.Join(absRoot, "docs", "examples", appDir)
	sourceTar, err := ioutil.TempFile(util.UnikTmpDir(), "")
	if err != nil {
		return nil, errors.New("failed to create tmp tar file", err)
	}
	defer os.Remove(sourceTar.Name())
	if err := unikos.Compress(path, sourceTar.Name()); err != nil {
		return nil, errors.New("failed to tar sources", err)
	}
	return sourceTar, nil
}

func BuildExampleImage(daemonUrl, projectRoot, exampleName, compiler, provider string, mounts []string) (*types.Image, error) {
	force := false
	noCleanup := false
	testSourceTar, err := TarExampleApp(projectRoot, exampleName)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(testSourceTar.Name())
	return client.UnikClient(daemonUrl).Images().Build(exampleName, testSourceTar.Name(), compiler, provider, "", mounts, force, noCleanup)
}

func RunExampleInstance(daemonUrl, instanceName, imageName string, volsToMounts map[string]string) (*types.Instance, error) {
	noCleanup := false
	env := map[string]string{"FOO": "BAR"}
	memoryMb := 128
	return client.UnikClient(daemonUrl).Instances().Run(instanceName, imageName, volsToMounts, env, memoryMb, noCleanup)
}

func CreateExampleVolume(daemonUrl, volumeName, provider string, size int) (*types.Volume, error) {
	return client.UnikClient(daemonUrl).Volumes().Create(volumeName, "", provider, size, false)
}

func GetProjectRoot() string {
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		_, filename, _, ok := runtime.Caller(1)
		if !ok {
			logrus.Panic("could not get current file")
		}
		projectRoot = filepath.Join(filepath.Dir(filename), "..", "..")
	}
	return projectRoot
}