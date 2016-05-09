package rump

import (
	"fmt"

	"github.com/emc-advanced-dev/unik/pkg/types"
)

func CreateImageAws(kernel string, args string, mntPoints []string) (*types.RawImage, error) {

	// create rump config
	var c rumpConfig

	if args == "" {
		c.Cmdline = "program.bin"
	} else {
		c.Cmdline = "program.bin" + " " + args
	}

	res := &types.RawImage{}
	volIndex := 0
	// add root -> sda1 mapping
	res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings, types.DeviceMapping{MountPoint: "/", DeviceName: "/dev/sda1"})

	for _, mntPoint := range mntPoints {
		// start from sdb; sda is for root.
		volIndex++
		deviceMapped := fmt.Sprintf("sd%c1", 'a'+volIndex)
		blk := blk{
			Source:     "etfs",
			Path:       deviceMapped,
			FSType:     "blk",
			MountPoint: mntPoint,
		}

		c.Blk = append(c.Blk, blk)
		res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings, types.DeviceMapping{MountPoint: mntPoint, DeviceName: "/dev/" + deviceMapped})
	}

	// aws network
	c.Net = &net{
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
	res.StageSpec = types.StageSpec{
		ImageFormat: types.ImageFormat_RAW,
		XenVirtualizationType: types.XenVirtualizationType_Paravirtual,
	}
	res.RunSpec.DefaultInstanceMemory = 1024
	
	return res, nil

}
