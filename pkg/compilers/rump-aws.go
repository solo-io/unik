package compilers

import (
	"fmt"

	uniktypes "github.com/emc-advanced-dev/unik/pkg/types"
)

func CreateImageAws(kernel string, args string, mntPoints []string) (*uniktypes.RawImage, error) {

	// create rump config
	var c RumpConfig

	if args == "" {
		c.Cmdline = "program.bin"
	} else {
		c.Cmdline = "program.bin" + " " + args
	}

	res := &uniktypes.RawImage{}
	volIndex := 0
	// add root -> sda1 mapping
	res.DeviceMappings = append(res.DeviceMappings, &uniktypes.DeviceMapping{MountPoint: "/", DeviceName: "/dev/sda1"})

	for _, mntPoint := range mntPoints {
		// start from sdb; sda is for root.
		volIndex++
		deviceMapped := fmt.Sprintf("sd%c1", 'a'+volIndex)
		blk := Blk{
			Source:     "etfs",
			Path:       deviceMapped,
			FSType:     "blk",
			MountPoint: mntPoint,
		}

		c.Blk = append(c.Blk, blk)
		res.DeviceMappings = append(res.DeviceMappings, &uniktypes.DeviceMapping{MountPoint: mntPoint, DeviceName: "/dev/" + deviceMapped})
	}

	// aws network
	c.Net = &Net{
		If:     "xenif0",
		Cloner: "true",
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
