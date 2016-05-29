package rump

import (
	"fmt"
	"strings"

	"github.com/emc-advanced-dev/unik/pkg/types"
)

func CreateImageAws(kernel string, args string, mntPoints, bakedEnv []string, noCleanup bool) (*types.RawImage, error) {

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

	cmdline, err := toRumpJson(c)
	if err != nil {
		return nil, err
	}
	imgFile, err := BuildBootableImage(kernel, cmdline, noCleanup)

	if err != nil {
		return nil, err
	}

	res.LocalImagePath = imgFile
	res.StageSpec = types.StageSpec{
		ImageFormat:           types.ImageFormat_RAW,
		XenVirtualizationType: types.XenVirtualizationType_Paravirtual,
	}
	res.RunSpec.DefaultInstanceMemory = 1024

	return res, nil

}
