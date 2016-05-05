package common

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"os/exec"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"encoding/json"
)

func ConvertRawImage(sourceFormat, targetFormat types.ImageFormat, inputFile, outputFile string) error {
	targetFormatName := string(targetFormat)
	if targetFormat == types.ImageFormat_VHD {
		targetFormatName = "vpc" //for some reason qemu calls VHD disks vpc
	}
	cmd := exec.Command("qemu-img", "convert", "-f", string(sourceFormat), "-O", targetFormatName, inputFile, outputFile)
	logrus.WithField("command", cmd.Args).Debugf("running command")
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.New("failed converting raw image to "+string(targetFormat)+": "+string(out), err)
	}
	return nil
}

func GetVirtualImageSize(imageFile string, imageFormat types.ImageFormat) (int64, error) {
	formatName := string(imageFormat)
	if imageFormat == types.ImageFormat_VHD {
		formatName = "vpc" //for some reason qemu calls VHD disks vpc
	}
	cmd := exec.Command("qemu-img", "info", "--output", "json", "-f", formatName, imageFile)
	logrus.WithField("command", cmd.Args).Debugf("running command")
	out, err := cmd.CombinedOutput()
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
	VirtualSize int64 `json:"virtual-size"`
	Filename string `json:"filename"`
	ClusterSize int `json:"cluster-size"`
	Format string `json:"format"`
	ActualSize int `json:"actual-size"`
	DirtyFlag bool `json:"dirty-flag"`
}