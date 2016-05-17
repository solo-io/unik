package rump

import (
	"fmt"

	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/Sirupsen/logrus"
)

func CreateImageVmware(kernel string, args string, mntPoints, bakedEnv []string) (*types.RawImage, error) {

	// create rump config
	var c rumpConfig
	if bakedEnv != nil {
		c.Env = bakedEnv
	}

	if args == "" {
		c.Cmdline = "program.bin"
	} else {
		c.Cmdline = "program.bin" + " " + args
	}

	res := &types.RawImage{}
	// add root -> sd0 mapping
	res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings, types.DeviceMapping{MountPoint: "/", DeviceName: "sd0"})

	for i, mntPoint := range mntPoints {
		deviceMapped := fmt.Sprintf("sd1%c", 'a'+i)
		blk := blk{
			Source:     "dev",
			Path:       "/dev/" + deviceMapped,
			FSType:     "blk",
			MountPoint: mntPoint,
		}

		c.Blk = append(c.Blk, blk)
		res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings, types.DeviceMapping{MountPoint: mntPoint, DeviceName: deviceMapped})
	}

	// aws network
	c.Net = &net{
		If:     "wm0",
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
	res.StageSpec.ImageFormat = types.ImageFormat_RAW
	res.RunSpec.StorageDriver = types.StorageDriver_SCSI
	res.RunSpec.VsphereNetworkType = types.VsphereNetworkType_E1000
	res.RunSpec.DefaultInstanceMemory = 512
	logrus.WithField("runspec", res.RunSpec).Infof("created raw vmware image")
	return res, nil

}
