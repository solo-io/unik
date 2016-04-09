package compilers

import (
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	uniktypes "github.com/emc-advanced-dev/unik/pkg/types"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/engine-api/types/network"

	"golang.org/x/net/context"

	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"path"
	"os"
)

// uses rump docker conter container
// the container expectes code in /opt/code and will produce program.bin in the same folder.
// we need to take the program bin and combine with json config produce an image

type RunmpCompiler struct {
	DockerImage string
	CreateImage func(kernel string, mntPoints []string) (*uniktypes.RawImage, error)
}

func (r *RunmpCompiler) CompileRawImage(sourceTar multipart.File, tarHeader *multipart.FileHeader, mntPoints []string) (*uniktypes.RawImage, error) {

	localFolder, err := ioutil.TempDir("","")
	if err != nil {
		return nil, err
	}
    defer os.RemoveAll(localFolder)
   
   // TODO: need to extract tar file there..
   
	err = r.runContainer(localFolder)
	if err != nil {
		return nil, err
	}

	// now we should program.bin

	resultFile := path.Join(localFolder, "program.bin")

	return r.CreateImage(resultFile, mntPoints)
}

func (r *RunmpCompiler) runContainer(localFolder string) error {

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	binds := []string{fmt.Sprintf("%s:%s", localFolder, "/opt/code")}

	config := &container.Config{
		Image: r.DockerImage,
	}
	var hostConfig *container.HostConfig = &container.HostConfig{
		Binds: binds,
	}
	var networkingConfig *network.NetworkingConfig = &network.NetworkingConfig{}
	containerName := ""
	container, err := cli.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, containerName)

	if err != nil {
		log.WithField("err", err).Error("Error creating container")
		return err
	}

	defer cli.ContainerRemove(context.Background(), types.ContainerRemoveOptions{ContainerID: container.ID})

	err = cli.ContainerStart(context.Background(), container.ID)

	if err != nil {
		log.WithField("err", err).Error("ContainerStart")
		return err
	}

	status, err := cli.ContainerWait(context.Background(), container.ID)
	if err != nil {
		return err
	}

	if status != 0 {

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

		res, _ := ioutil.ReadAll(reader)
		fmt.Print("output:")
		fmt.Print(string(res))

		return errors.New("Returned non zero status")
	}

	return nil
}
