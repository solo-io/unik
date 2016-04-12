package os

import (
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/engine-api/types/strslice"
	"golang.org/x/net/context"
	"github.com/layer-x/layerx-commons/lxlog"
	"io/ioutil"
	"github.com/layer-x/layerx-commons/lxerrors"
)


func RunContainer(logger lxlog.Logger, imageName string, cmds, binds []string, privileged bool) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	config := &container.Config{
		Image: imageName,
		Cmd:   strslice.StrSlice(cmds),
	}
	hostConfig := &container.HostConfig{
		Binds:      binds,
		Privileged: privileged,
	}
	networkingConfig := &network.NetworkingConfig{}
	containerName := ""

	container, err := cli.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, containerName)
	if err != nil {
		logger.WithErr(err).Errorf("Error creating container")
		return err
	}
	defer cli.ContainerRemove(context.Background(), types.ContainerRemoveOptions{ContainerID: container.ID})

	logger.WithField("id", container.ID).Errorf("Created container")

	if err := cli.ContainerStart(context.Background(), container.ID); err != nil {
		logger.WithErr(err).Errorf("ContainerStart")
		return err
	}

	status, err := cli.ContainerWait(context.Background(), container.ID)
	if err != nil {
		return err
	}

	if status != 0 {
		logger.WithField("status", status).Errorf("Container exit status non zero")

		options := types.ContainerLogsOptions{
			ContainerID: container.ID,
			ShowStdout:  true,
			ShowStderr:  true,
			Follow:      true,
			Tail:        "all",
		}
		reader, err := cli.ContainerLogs(context.Background(), options)
		if err != nil {
			logger.WithErr(err).Errorf("ContainerLogs")
			return err
		}
		defer reader.Close()

		if res, err := ioutil.ReadAll(reader); err == nil {
			logger.Errorf(string(res))
		} else {
			logger.WithErr(err).Warnf("failed to get logs")
		}

		return lxerrors.New("Returned non zero status", nil)
	}

	return nil
}