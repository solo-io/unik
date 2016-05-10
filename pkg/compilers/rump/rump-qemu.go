package rump

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/config"
	uniktypes "github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
)

func CreateImageQemu(kernel string, args string, mntPoints []string) (*uniktypes.RawImage, error) {

	// create rump config
	var c rumpConfig

	if args == "" {
		c = setRumpCmdLine(c, "program.bin", nil)
	} else {
		c = setRumpCmdLine(c, "program.bin", strings.Split(args, " "))
	}

	res := &uniktypes.RawImage{}
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
		res.DeviceMappings = append(res.DeviceMappings,
			uniktypes.DeviceMapping{MountPoint: mntPoint, DeviceName: deviceMapped})
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
	destZip, err := ioutil.TempFile(unikutil.UnikTmpDir(), "qemu_zip_")
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
