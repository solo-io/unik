package rump

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func CreateImageQemu(kernel string, args string, mntPoints, bakedEnv []string) (*types.RawImage, error) {

	// create rump config
	var c rumpConfig

	if bakedEnv != nil {
		c.Env = bakedEnv
	}

	if args == "" {
		c = setRumpCmdLine(c, "program.bin", nil, false)
	} else {
		c = setRumpCmdLine(c, "program.bin", strings.Split(args, " "), false)
	}

	res := &types.RawImage{}
	// add root -> sd0 mapping
	for i, mntPoint := range mntPoints {
		deviceMapped := fmt.Sprintf("ld%ca", '0'+i)
		blk := blk{
			Source:     "dev",
			Path:       "/dev/" + deviceMapped,
			FSType:     "blk",
			MountPoint: mntPoint,
		}

		c.Blk = append(c.Blk, blk)
		logrus.Debugf("adding mount point to image: %s:%s", mntPoint, deviceMapped)
		res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings,
			types.DeviceMapping{MountPoint: mntPoint, DeviceName: deviceMapped})
		res.RunSpec.Compiler = compilers.Rump
	}

	// virtualbox network
	c.Net = &net{
		If:     "vioif0",
		Type:   "inet",
		Method: DHCP,
	}

	cmdline, err := ToRumpJson(c)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("writing rump json config: %s", cmdline)

	imgFile, err := zipFiles(kernel, cmdline)
	if err != nil {
		return nil, err
	}

	res.LocalImagePath = imgFile
	return res, nil

}

func zipFiles(kernelFile string, cmdline string) (string, error) {
	destZip, err := ioutil.TempFile("", "TMPqemu_zip_")
	if err != nil {
		return "", err
	}
	defer destZip.Close()
	w := zip.NewWriter(destZip)

	kernelReader, err := os.Open(kernelFile)
	if err != nil {
		return "", err
	}
	defer kernelReader.Close()

	// create kernel file
	f, err := w.Create(config.QemuKernelFileName)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(f, kernelReader)
	if err != nil {
		return "", err
	}

	// create cmdline file
	f, err = w.Create(config.QemuArgsFileName)
	if err != nil {
		return "", err
	}
	_, err = f.Write([]byte(cmdline))
	if err != nil {
		return "", err
	}

	if err := w.Close(); err != nil {
		return "", err
	}

	return destZip.Name(), nil

}
