package os

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/emc-advanced-dev/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func BuildRawDataImage(dataTar io.ReadCloser, size int, usePartitionTables bool) (string, error) {
	dataFolder, err := ioutil.TempDir(unikutil.UnikTmpDir(), "")
	if err != nil {
		return "", errors.New("creating tmp build folder", err)
	}
	defer os.RemoveAll(dataFolder)

	if err := ExtractTar(dataTar, dataFolder); err != nil {
		return "", errors.New("extracting data tar", err)
	}

	buildDir := filepath.Dir(dataFolder)

	var cmd *exec.Cmd
	if size > 0 {
		cmd = exec.Command("docker", "run", "--rm", "--privileged",
			"-v", "/dev/:/dev/",
			"-v", buildDir+"/:/opt/vol/",
			"projectunik/image-creator",
			"-p", fmt.Sprintf("%v", usePartitionTables),
			"-v", filepath.Base(dataFolder), fmt.Sprintf(",%v", size),
		)
	} else {
		cmd = exec.Command("docker", "run", "--rm", "--privileged",
			"-v", "/dev/:/dev/",
			"-v", buildDir+"/:/opt/vol/",
			"projectunik/image-creator",
			"-p", fmt.Sprintf("%v", usePartitionTables),
			"-v", filepath.Base(dataFolder),
		)
	}

	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running image-creator container")
	unikutil.LogCommand(cmd, true)
	err = cmd.Run()
	if err != nil {
		return "", errors.New("failed running image-creator on "+dataFolder, err)
	}

	resultFile, err := ioutil.TempFile(unikutil.UnikTmpDir(), "")
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
		return "", errors.New("must specify size > 0", nil)
	}
	dataFolder, err := ioutil.TempDir(unikutil.UnikTmpDir(), "")
	if err != nil {
		return "", errors.New("creating tmp build folder", err)
	}
	defer os.RemoveAll(dataFolder)

	buildDir := filepath.Dir(dataFolder)

	cmd := exec.Command("docker", "run", "--rm", "--privileged",
		"-v", "/dev/:/dev/",
		"-v", buildDir+"/:/opt/vol/",
		"projectunik/image-creator",
		"-v", filepath.Base(dataFolder), fmt.Sprintf(",%v", size),
	)

	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running image-creator container")
	unikutil.LogCommand(cmd, true)
	err = cmd.Run()
	if err != nil {
		return "", errors.New("failed running image-creator on "+dataFolder, err)
	}

	resultFile, err := ioutil.TempFile(unikutil.UnikTmpDir(), "")
	if err != nil {
		return "", err
	}
	resultFile.Close()

	if err := os.Rename(path.Join(buildDir, "vol.img"), resultFile.Name()); err != nil {
		return "", err
	}

	return resultFile.Name(), nil
}
