package compilers

import (
	"archive/tar"
	"encoding/json"
	"io"

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
	"os"
	"path"
)

// uses rump docker conter container
// the container expectes code in /opt/code and will produce program.bin in the same folder.
// we need to take the program bin and combine with json config produce an image

type RunmpCompiler struct {
	DockerImage string
	CreateImage func(kernel, args string, mntPoints []string) (*uniktypes.RawImage, error)
}

func (r *RunmpCompiler) CompileRawImage(sourceTar io.ReadCloser, args string, mntPoints []string) (*uniktypes.RawImage, error) {

	localFolder, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(localFolder)

	if err := r.extractTar(sourceTar, localFolder); err != nil {
		return nil, err
	}

	if err := r.runContainer(localFolder); err != nil {
		return nil, err
	}

	// now we should program.bin
	resultFile := path.Join(localFolder, "program.bin")

	return r.CreateImage(resultFile, args, mntPoints)
}

func (r *RunmpCompiler) extractTar(tarArchive io.ReadCloser, localFolder string) error {
	tr := tar.NewReader(tarArchive)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return err
		}
		log.WithField("file", hdr.Name).Debug("Extracting file")
		switch hdr.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(path.Join(localFolder, hdr.Name), 0755)

			if err != nil {
				return err
			}

		case tar.TypeReg:
			fallthrough
		case tar.TypeRegA:
			dir, _ := path.Split(hdr.Name)
			if err := os.MkdirAll(path.Join(localFolder, dir), 0755); err != nil {
				return err
			}

			outputFile, err := os.Create(path.Join(localFolder, hdr.Name))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outputFile, tr); err != nil {
				outputFile.Close()
				return err
			}
			outputFile.Close()

		default:
			return errors.New("Unsupported file type in tar")
		}
	}

	return nil
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
	hostConfig := &container.HostConfig{
		Binds: binds,
	}
	networkingConfig := &network.NetworkingConfig{}
	containerName := ""

	container, err := cli.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, containerName)
	if err != nil {
		log.WithField("err", err).Error("Error creating container")
		return err
	}
	defer cli.ContainerRemove(context.Background(), types.ContainerRemoveOptions{ContainerID: container.ID})

	log.WithField("id", container.ID).Error("Created container")

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

		return errors.New("Returned non zero status")
	}

	return nil
}

// rump special json
func ToRumpJson(c RumpConfig) (string, error) {

	blk := c.Blk
	c.Blk = nil

	jsonConfig, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	blks := ""
	for _, b := range blk {

		blkjson, err := json.Marshal(b)
		if err != nil {
			return "", err
		}
		blks += fmt.Sprintf("\"blk\": %s,", string(blkjson))
	}
	var jsonString string
	if len(blks) > 0 {

		jsonString = string(jsonConfig[:len(jsonConfig)-1]) + "," + blks[:len(blks)-1] + "}"

	} else {
		jsonString = string(jsonConfig)
	}

	return jsonString, nil

}
