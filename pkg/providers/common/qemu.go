package common

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"io/ioutil"
)

func ConvertRawImage(sourceFormat, targetFormat types.ImageFormat, inputFile, outputFile string) error {
	targetFormatName := string(targetFormat)
	if targetFormat == types.ImageFormat_VHD {
		targetFormatName = "vpc" //for some reason qemu calls VHD disks vpc
	}
	dir := filepath.Dir(inputFile)
	outDir := filepath.Dir(outputFile)

	container := unikutil.NewContainer("qemu-util").WithVolume(dir, dir).
		WithVolume(outDir, outDir)

	args := []string{"qemu-img", "convert", "-f", string(sourceFormat), "-O", targetFormatName}
	if targetFormat == types.ImageFormat_VMDK {
		args = append(args, "-o", "compat6")
	}

	//this needs to be done because docker produces files as root. argh!!!
	tmpOutputFile, err := ioutil.TempFile(outDir, "convert.image.result.")
	if err != nil {
		return errors.New("temp file for root user", err)
	}
	tmpOutputFile.Close()
	defer os.Remove(tmpOutputFile.Name())

	args = append(args, inputFile, tmpOutputFile.Name())

	logrus.WithField("command", args).Debugf("running command")
	if err := container.Run(args...); err != nil {
		return errors.New("failed converting raw image to "+string(targetFormat), err)
	}

	if err := unikos.CopyFile(tmpOutputFile.Name(), outputFile); err != nil {
		return errors.New("copying tmp result to final result", err)
	}

	return nil
}

func fixVmdk(vmdkFile string) error {
	file, err := os.OpenFile(vmdkFile, os.O_RDWR, 0)
	if err != nil {
		return errors.New("can't open vmdk", err)
	}
	defer file.Close()

	var buffer [1024]byte

	n, err := file.Read(buffer[:])
	if err != nil {
		return errors.New("can't read vmdk", err)
	}
	if n < len(buffer) {
		return errors.New("bad vmdk", err)
	}

	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		return errors.New("can't seek vmdk", err)
	}

	result := bytes.Replace(buffer[:], []byte("# The disk Data Base"), []byte("# The Disk Data Base"), 1)

	_, err = file.Write(result)
	if err != nil {
		return errors.New("can't write vmdk", err)
	}

	return nil
}

func ConvertRawToNewVmdk(inputFile, outputFile string) error {

	dir := filepath.Dir(inputFile)
	outDir := filepath.Dir(outputFile)

	container := unikutil.NewContainer("euranova/ubuntu-vbox").WithVolume(dir, dir).
		WithVolume(outDir, outDir)

	args := []string{
		"VBoxManage", "convertfromraw", inputFile, outputFile, "--format", "vmdk", "--variant", "Stream"}

	logrus.WithField("command", args).Debugf("running command")
	if err := container.Run(args...); err != nil {
		return errors.New("failed converting raw image to vmdk", err)
	}

	err := fixVmdk(outputFile)
	if err != nil {
		return errors.New("failed converting raw image to vmdk", err)
	}

	return nil
}

func GetVirtualImageSize(imageFile string, imageFormat types.ImageFormat) (int64, error) {
	formatName := string(imageFormat)
	if imageFormat == types.ImageFormat_VHD {
		formatName = "vpc" //for some reason qemu calls VHD disks vpc
	}
	dir := filepath.Dir(imageFile)

	container := unikutil.NewContainer("qemu-util").WithVolume(dir, dir)
	args := []string{"qemu-img", "info", "--output", "json", "-f", formatName, imageFile}

	logrus.WithField("command", args).Debugf("running command")
	out, err := container.CombinedOutput(args...)
	if err != nil {
		return -1, errors.New("failed getting image info", err)
	}
	var info imageInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return -1, errors.New("parsing "+string(out)+" to json", err)
	}
	return info.VirtualSize, nil
}

type imageInfo struct {
	VirtualSize int64  `json:"virtual-size"`
	Filename    string `json:"filename"`
	ClusterSize int    `json:"cluster-size"`
	Format      string `json:"format"`
	ActualSize  int    `json:"actual-size"`
	DirtyFlag   bool   `json:"dirty-flag"`
}
