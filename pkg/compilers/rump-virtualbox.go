package compilers

import (
	"fmt"

	uniktypes "github.com/emc-advanced-dev/unik/pkg/types"
	"regexp"
	"github.com/Sirupsen/logrus"
)

func CreateImageVirtualBox(kernel string, args string, mntPoints []string) (*uniktypes.RawImage, error) {

	// create rump config
	var c multinetRumpConfig

	if args == "" {
		c.Cmdline = "program.bin"
	} else {
		c.Cmdline = "program.bin" + " " + args
	}

	res := &uniktypes.RawImage{}
	// add root -> sd0 mapping
	res.DeviceMappings = append(res.DeviceMappings,
		uniktypes.DeviceMapping{MountPoint: "/", DeviceName: "sd0"})

	for i, mntPoint := range mntPoints {
		deviceMapped := fmt.Sprintf("sd%ca", '1'+i)
		blk := blk{
			Source:     "dev",
			Path:       "/dev/" + deviceMapped,
			FSType:     "blk",
			MountPoint: mntPoint,
		}

		c.Blk = append(c.Blk, blk)
		res.DeviceMappings = append(res.DeviceMappings,
			uniktypes.DeviceMapping{MountPoint: mntPoint, DeviceName: deviceMapped})
	}

	// virtualbox network
	c.Net1 = &net{
		If:     "vioif0",
		Type:   "inet",
		Method: DHCP,
	}
	c.Net2 = &net{
		If:     "vioif1",
		Type:   "inet",
		Method: DHCP,
	}

	cmdline, err := ToRumpJsonMultiNet(c)
	if err != nil {
		return nil, err
	}

	r, err := regexp.Compile("net[1-9]")
	if err != nil {
		return nil, err
	}
	cmdline = string(r.ReplaceAll([]byte(cmdline), []byte("net")))

	logrus.Debugf("writing rump json config: %s", cmdline)

	imgFile, err := BuildBootableImage(kernel, cmdline)
	if err != nil {
		return nil, err
	}

	res.LocalImagePath = imgFile
	return res, nil

}
