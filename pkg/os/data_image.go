package os

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
)

func BuildRawDataImage(dataTar io.ReadCloser, size MegaBytes, usePartitionTables bool) (string, error) {
	dataFolder, err := ioutil.TempDir(unikutil.UnikTmpDir(), "")
	if err != nil {
		return "", errors.New("creating tmp build folder", err)
	}
	defer os.RemoveAll(dataFolder)

	if err := ExtractTar(dataTar, dataFolder); err != nil {
		return "", errors.New("extracting data tar", err)
	}

	buildDir := filepath.Dir(dataFolder)

	container := unikutil.NewContainer("image-creator").Privileged(true).WithVolume("/dev/", "/dev/").
		WithVolume(buildDir+"/", "/opt/vol")

	var args []string
	if size > 0 {
		args = append(args, "-p", fmt.Sprintf("%v", usePartitionTables),
			"-v", fmt.Sprintf("%s,%v", filepath.Base(dataFolder), size.ToBytes()))
	} else {
		args = append(args, "-p", fmt.Sprintf("%v", usePartitionTables),
			"-v", filepath.Base(dataFolder),
		)
	}

	logrus.WithFields(logrus.Fields{
		"command": args,
	}).Debugf("running image-creator container")

	if err = container.Run(args...); err != nil {
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

func BuildEmptyDataVolume(size MegaBytes) (string, error) {
	if size < 1 {
		return "", errors.New("must specify size > 0", nil)
	}
	dataFolder, err := ioutil.TempDir(unikutil.UnikTmpDir(), "")
	if err != nil {
		return "", errors.New("creating tmp build folder", err)
	}
	defer os.RemoveAll(dataFolder)

	buildDir := filepath.Dir(dataFolder)

	container := unikutil.NewContainer("image-creator").Privileged(true).WithVolume("/dev/", "/dev/").
		WithVolume(buildDir+"/", "/opt/vol")

	args := []string{"-v", fmt.Sprintf("%s,%v", filepath.Base(dataFolder), size.ToBytes())}

	logrus.WithFields(logrus.Fields{
		"command": args,
	}).Debugf("running image-creator container")
	if err := container.Run(args...); err != nil {
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
