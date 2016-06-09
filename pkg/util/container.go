package util

import (
	"fmt"
	"os/exec"
)

func SetContainerVer(ver string) {
	containerVer = ver
}

// filled in build time by make
var containerVer string

type Container struct {
	env        map[string]string
	privileged bool
	volumes    map[string]string
	name       string
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
func (c *Container) Privileged(p bool) *Container {
	c.privileged = p
	return c
}

func (c *Container) Run(arguments ...string) error {
	cmd := c.buildCmd(arguments...)

	LogCommand(cmd, true)

	return cmd.Run()
}

func (c *Container) Output(arguments ...string) ([]byte, error) {
	return c.buildCmd(arguments...).Output()
}

func (c *Container) CombinedOutput(arguments ...string) ([]byte, error) {
	return c.buildCmd(arguments...).CombinedOutput()
}

func (c *Container) buildCmd(arguments ...string) *exec.Cmd {
	args := []string{"run", "--rm"}
	if c.privileged {
		args = append(args, "--privileged")
	}
	for key, val := range c.env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, val))
	}
	for key, val := range c.volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", key, val))
	}

	args = append(args, "projectunik/"+c.name+":"+containerVer)
	args = append(args, arguments...)

	cmd := exec.Command("docker", args...)

	return cmd
}
