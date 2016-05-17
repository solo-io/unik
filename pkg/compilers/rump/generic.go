package rump

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/engine-api/types/strslice"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"golang.org/x/net/context"
	"os/exec"
	"strings"
)

func BuildBootableImage(kernel, cmdline string) (string, error) {
	directory, err := ioutil.TempDir(unikutil.UnikTmpDir(), "")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(directory)
	kernelBaseName := "program.bin"

	if err := unikos.CopyFile(kernel, path.Join(directory, kernelBaseName)); err != nil {
		return "", err
	}

	const contextDir = "/opt/vol/"
	cmds := []string{"-d", contextDir, "-p", kernelBaseName, "-a", cmdline}
	binds := []string{directory + ":" + contextDir, "/dev/:/dev/"}

	if err := execContainer("projectunik/boot-creator", cmds, binds, true, nil); err != nil {
		return "", err
	}

	resultFile, err := ioutil.TempFile(unikutil.UnikTmpDir(), "")
	if err != nil {
		return "", err
	}
	resultFile.Close()

	if err := os.Rename(path.Join(directory, "vol.img"), resultFile.Name()); err != nil {
		return "", err
	}

	return resultFile.Name(), nil
}

func RunContainer(imageName string, cmds, binds []string, privileged bool, envPairs []string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	config := &container.Config{
		Image: imageName,
		Cmd:   strslice.StrSlice(cmds),
		Env: envPairs,
	}
	hostConfig := &container.HostConfig{
		Binds:      binds,
		Privileged: privileged,
	}
	networkingConfig := &network.NetworkingConfig{}
	containerName := ""

	container, err := cli.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, containerName)
	if err != nil {
		log.WithField("err", err).Error("Error creating container")
		return err
	}
	defer cli.ContainerRemove(context.Background(), types.ContainerRemoveOptions{ContainerID: container.ID})

	log.WithFields(log.Fields{"id": container.ID, "cmd": cmds, "binds": binds}).Info("Created container")

	if err := cli.ContainerStart(context.Background(), container.ID); err != nil {
		log.WithField("err", err).Error("ContainerStart")
		return err
	}

	status, err := cli.ContainerWait(context.Background(), container.ID)
	if err != nil {
		return err
	}

	if status != 0 {
		log.WithField("status", status).Error("Container exit status non zero")

		options := types.ContainerLogsOptions{
			ContainerID: container.ID,
			ShowStdout:  true,
			ShowStderr:  true,
			Follow:      true,
			Tail:        "all",
		}
		reader, err := cli.ContainerLogs(context.Background(), options)
		if err != nil {
			log.WithField("err", err).Error("ContainerLogs")
			return err
		}
		defer reader.Close()

		if res, err := ioutil.ReadAll(reader); err == nil {
			log.Error(string(res))
		} else {
			log.WithField("err", err).Warn("failed to get logs")
		}

		return errors.New("Returned non zero status", nil)
	}

	return nil
}

func execContainer(imageName string, cmds, binds []string, privileged bool, env map[string]string) error {
	dockerArgs := []string{"run", "--rm"}
	if privileged {
		dockerArgs = append(dockerArgs, "--privileged")
	}
	for _, bind := range binds {
		dockerArgs = append(dockerArgs, "-v", bind)
	}
	for key, val := range env {
		dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", key, val))
	}
	dockerArgs = append(dockerArgs, imageName)
	dockerArgs = append(dockerArgs, cmds...)
	cmd := exec.Command("docker", dockerArgs...)
	unikutil.LogCommand(cmd, true)
	if err := cmd.Run(); err != nil {
		return errors.New("running container "+imageName, err)
	}
	return nil
}

func (r *RumpCompiler) runContainer(localFolder string, envPairs []string) error {
	//return RunContainer(r.DockerImage, nil, []string{fmt.Sprintf("%s:%s", localFolder, "/opt/code")}, false, envPairs)
	env := make(map[string]string)
	for _, pair := range envPairs {
		split := strings.Split(pair, "=")
		if len(split) != 2 {
			return errors.New(pair+" is invaid string for env pair", nil)
		}
		env[split[0]] = split[1]
	}
	return execContainer(r.DockerImage, nil, []string{fmt.Sprintf("%s:%s", localFolder, "/opt/code")}, false, env)
}
