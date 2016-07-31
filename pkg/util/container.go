// +build !container-binary

package util

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/containers"
	"github.com/pborman/uuid"
)

var containerVersions map[string]string

func InitContainers() error {
	versionData, err := versiondata.Asset("containers/versions.json")
	if err != nil {
		return errors.New("failed to get version data from containers/versions.json", err)
	}
	if err := json.Unmarshal(versionData, &containerVersions); err != nil {
		return errors.New("failed to unmarshall version data "+string(versionData), err)
	}
	logrus.WithField("versions", containerVersions).Info("using container versions")
	return nil
}

type Container struct {
	env           map[string]string
	privileged    bool
	volumes       map[string]string
	interactive   bool
	network       string
	containerName string
	name          string
}

func NewContainer(imageName string) *Container {
	c := &Container{}

	c.name = imageName
	c.env = make(map[string]string)
	c.volumes = make(map[string]string)

	return c
}

func (c *Container) WithVolume(hostdir, containerdir string) *Container {
	c.volumes[hostdir] = containerdir
	return c
}

func (c *Container) WithVolumes(vols map[string]string) *Container {
	for k, v := range vols {
		c.WithVolume(k, v)
	}
	return c
}

func (c *Container) WithEnv(key, value string) *Container {
	c.env[key] = value
	return c
}

func (c *Container) WithEnvs(vars map[string]string) *Container {
	for k, v := range vars {
		c.WithEnv(k, v)
	}
	return c
}

func (c *Container) WithNet(net string) *Container {
	c.network = net
	return c
}

func (c *Container) WithName(name string) *Container {
	c.containerName = name
	return c
}

func (c *Container) Interactive(i bool) *Container {
	c.interactive = i
	return c
}

func (c *Container) Privileged(p bool) *Container {
	c.privileged = p
	return c
}

func (c *Container) Run(arguments ...string) error {
	cmd := c.BuildCmd(arguments...)

	LogCommand(cmd, true)

	return cmd.Run()
}

func (c *Container) Output(arguments ...string) ([]byte, error) {
	return c.BuildCmd(arguments...).Output()
}

func (c *Container) CombinedOutput(arguments ...string) ([]byte, error) {
	return c.BuildCmd(arguments...).CombinedOutput()
}

func (c *Container) Stop() error {
	return exec.Command("docker", "stop", c.containerName).Run()
}

func (c *Container) BuildCmd(arguments ...string) *exec.Cmd {
	if c.containerName == "" {
		c.containerName = uuid.New()
	}

	args := []string{"run", "--rm"}
	if c.privileged {
		args = append(args, "--privileged")
	}
	if c.interactive {
		args = append(args, "-i")
	}
	if c.network != "" {
		args = append(args, fmt.Sprintf("--net=%s", c.network))
	}
	for key, val := range c.env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, val))
	}
	for key, val := range c.volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", key, val))
	}

	args = append(args, fmt.Sprintf("--name=%s", c.containerName))

	containerVer, ok := containerVersions[c.name]
	if !ok {
		logrus.Warnf("version for container %s not found, using version 'latest'", c.name)
		containerVer = "latest"
	}

	finalName := "projectunik/" + c.name + ":" + containerVer
	logrus.Info("Starting container ", finalName)

	args = append(args, finalName)
	args = append(args, arguments...)

	cmd := exec.Command("docker", args...)

	return cmd
}
