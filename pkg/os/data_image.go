package os

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	uniklog "github.com/emc-advanced-dev/unik/pkg/util/log"
	"github.com/layer-x/layerx-commons/lxerrors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func BuildRawDataImage(dataTar io.ReadCloser, size int, usePartitionTables bool) (string, error) {
	dataFolder, err := ioutil.TempDir("", "")
	if err != nil {
		return "", lxerrors.New("creating tmp build folder", err)
	}
	defer os.RemoveAll(dataFolder)

	if err := ExtractTar(dataTar, dataFolder); err != nil {
		return "", lxerrors.New("extracting data tar", err)
	}

	buildDir := filepath.Dir(dataFolder)

	var cmd *exec.Cmd
	if size > 0 {
		cmd = exec.Command("docker", "run", "--rm", "--privileged",
			"-v", "/dev/:/dev/",
			"-v", buildDir+"/:/opt/vol/",
			"unik/image-creator",
			"-p", fmt.Sprintf("%v", usePartitionTables),
			"-v", filepath.Base(dataFolder), fmt.Sprintf(",%v", size),
		)
	} else {
		cmd = exec.Command("docker", "run", "--rm", "--privileged",
			"-v", "/dev/:/dev/",
			"-v", buildDir+"/:/opt/vol/",
			"unik/image-creator",
			"-p", fmt.Sprintf("%v", usePartitionTables),
			"-v", filepath.Base(dataFolder),
		)
	}

	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running image-creator container")
	uniklog.LogCommand(cmd, true)
	err = cmd.Run()
	if err != nil {
		return "", lxerrors.New("failed running image-creator on "+dataFolder, err)
	}

	resultFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	resultFile.Close()

	if err := os.Rename(path.Join(buildDir, "vol.img"), resultFile.Name()); err != nil {
		return "", err
	}

	return resultFile.Name(), nil
}

func BuildEmptyDataVolume(size int) (string, error) {
	if size < 1 {
		return "", lxerrors.New("must specify size > 0", nil)
	}
	dataFolder, err := ioutil.TempDir("", "")
	if err != nil {
		return "", lxerrors.New("creating tmp build folder", err)
	}
	defer os.RemoveAll(dataFolder)

	buildDir := filepath.Dir(dataFolder)

	cmd := exec.Command("docker", "run", "--rm", "--privileged",
		"-v", "/dev/:/dev/",
		"-v", buildDir+"/:/opt/vol/",
		"unik/image-creator",
		"-v", filepath.Base(dataFolder), fmt.Sprintf(",%v", size),
	)

	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running image-creator container")
	uniklog.LogCommand(cmd, true)
	err = cmd.Run()
	if err != nil {
		return "", lxerrors.New("failed running image-creator on "+dataFolder, err)
	}

	resultFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	resultFile.Close()

	if err := os.Rename(path.Join(buildDir, "vol.img"), resultFile.Name()); err != nil {
		return "", err
	}

	return resultFile.Name(), nil
}
