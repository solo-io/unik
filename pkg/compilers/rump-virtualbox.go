package compilers

import (
	"fmt"

	uniktypes "github.com/emc-advanced-dev/unik/pkg/types"
)

func CreateImageVirtualBox(kernel string, args string, mntPoints []string) (*uniktypes.RawImage, error) {

	// create rump config
	var c RumpConfig

	if args == "" {
		c.Cmdline = "program.bin"
	} else {
		c.Cmdline = "program.bin" + " " + args
	}

	res := &uniktypes.RawImage{}
	// add root -> sd0 mapping
	res.DeviceMappings = append(res.DeviceMappings, &uniktypes.DeviceMapping{MountPoint: "/", DeviceName: "sd0"})

	for i, mntPoint := range mntPoints {
		deviceMapped := fmt.Sprintf("sd%ca", '1'+i)
		blk := Blk{
			Source:     "dev",
			Path:       "/dev/" + deviceMapped,
			FSType:     "blk",
			MountPoint: mntPoint,
		}

		c.Blk = append(c.Blk, blk)
		res.DeviceMappings = append(res.DeviceMappings, &uniktypes.DeviceMapping{MountPoint: mntPoint, DeviceName: deviceMapped})
	}

	// aws network
	c.Net = &Net{
		If:     "vioif0",
		Type:   "inet",
		Method: DHCP,
	}

	cmdline, err := ToRumpJson(c)
	if err != nil {
		return nil, err
	}

	imgFile, err := BuildBootableImage(kernel, cmdline)
	if err != nil {
		return nil, err
	}

	res.LocalImagePath = imgFile
	return res, nil

}
