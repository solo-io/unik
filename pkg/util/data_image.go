// +build !container-binary

package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	unikos "github.com/solo-io/unik/pkg/os"
)

func BuildRawDataImageWithType(dataTar io.ReadCloser, size unikos.MegaBytes, volType string, usePartitionTables bool) (string, error) {
	buildDir, err := ioutil.TempDir("", ".raw_data_image_folder.")
	if err != nil {
		return "", errors.New("creating tmp build folder", err)
	}
	defer os.RemoveAll(buildDir)

	dataFolder := filepath.Join(buildDir, "data")
	err = os.Mkdir(dataFolder, 0755)
	if err != nil {
		return "", errors.New("creating tmp data folder", err)
	}

	if err := unikos.ExtractTar(dataTar, dataFolder); err != nil {
		return "", errors.New("extracting data tar", err)
	}

	container := NewContainer("image-creator").Privileged(true).WithVolume("/dev/", "/dev/").
		WithVolume(buildDir+"/", "/opt/vol")

	tmpResultFile, err := ioutil.TempFile(buildDir, "data.image.result.img.")
	if err != nil {
		return "", err
	}
	tmpResultFile.Close()
	args := []string{"-o", filepath.Base(tmpResultFile.Name())}

	if size > 0 {
		args = append(args, "-p", fmt.Sprintf("%v", usePartitionTables),
			"-v", fmt.Sprintf("%s,%v", filepath.Base(dataFolder), size.ToBytes()))
	} else {
		args = append(args, "-p", fmt.Sprintf("%v", usePartitionTables),
			"-v", filepath.Base(dataFolder),
		)
	}
	args = append(args, "-t", volType)

	logrus.WithFields(logrus.Fields{
		"command": args,
	}).Debugf("running image-creator container")

	if err = container.Run(args...); err != nil {
		return "", errors.New("failed running image-creator on "+dataFolder, err)
	}

	resultFile, err := ioutil.TempFile("", "data-volume-creator-result.img.")
	if err != nil {
		return "", err
	}
	resultFile.Close()
	if err := os.Rename(tmpResultFile.Name(), resultFile.Name()); err != nil {
		return "", errors.New("renaming "+tmpResultFile.Name()+" to "+resultFile.Name(), err)
	}

	return resultFile.Name(), nil
}

func BuildRawDataImage(dataTar io.ReadCloser, size unikos.MegaBytes, usePartitionTables bool) (string, error) {
	return BuildRawDataImageWithType(dataTar, size, "ext2", usePartitionTables)
}
func BuildEmptyDataVolumeWithType(size unikos.MegaBytes, volType string) (string, error) {

	if size < 1 {
		return "", errors.New("must specify size > 0", nil)
	}
	dataFolder, err := ioutil.TempDir("", "empty.data.folder.")
	if err != nil {
		return "", errors.New("creating tmp build folder", err)
	}
	defer os.RemoveAll(dataFolder)

	buildDir := filepath.Dir(dataFolder)

	container := NewContainer("image-creator").Privileged(true).WithVolume("/dev/", "/dev/").
		WithVolume(buildDir+"/", "/opt/vol")

	tmpResultFile, err := ioutil.TempFile(buildDir, "data.image.result.img.")
	if err != nil {
		return "", err
	}
	tmpResultFile.Close()
	args := []string{"-v", fmt.Sprintf("%s,%v", filepath.Base(dataFolder), size.ToBytes()), "-o", filepath.Base(tmpResultFile.Name())}
	args = append(args, "-t", volType)

	logrus.WithFields(logrus.Fields{
		"command": args,
	}).Debugf("running image-creator container")
	if err := container.Run(args...); err != nil {
		return "", errors.New("failed running image-creator on "+dataFolder, err)
	}

	resultFile, err := ioutil.TempFile("", "empty-data-volume-creator-result.img.")
	if err != nil {
		return "", err
	}
	resultFile.Close()
	if err := os.Rename(tmpResultFile.Name(), resultFile.Name()); err != nil {
		return "", errors.New("renaming "+tmpResultFile.Name()+" to "+resultFile.Name(), err)
	}

	return resultFile.Name(), nil
}

func BuildEmptyDataVolume(size unikos.MegaBytes) (string, error) {
	return BuildEmptyDataVolumeWithType(size, "ext2")
}
