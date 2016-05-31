package rump

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func CreateImageVmware(kernel string, args string, mntPoints, bakedEnv []string, noCleanup bool) (*types.RawImage, error) {
	return createImageVmware(kernel, args, mntPoints, bakedEnv, false, noCleanup)
}

func CreateImageNoStubVmware(kernel string, args string, mntPoints, bakedEnv []string, noCleanup bool) (*types.RawImage, error) {
	return createImageVmware(kernel, args, mntPoints, bakedEnv, true, noCleanup)
}

func createImageVmware(kernel string, args string, mntPoints, bakedEnv []string, noStub, noCleanup bool) (*types.RawImage, error) {
	// create rump config
	var c rumpConfig
	if bakedEnv != nil {
		c.Env = make(map[string]string)
		for i, pair := range bakedEnv {
			c.Env[fmt.Sprintf("env%d", i)] = pair
		}
	}

	if args == "" {
		c = setRumpCmdLine(c, "program.bin", nil, noStub)
	} else {
		c = setRumpCmdLine(c, "program.bin", strings.Split(args, " "), noStub)
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

	cmdline, err := toRumpJson(c)
	if err != nil {
		return nil, err
	}

	imgFile, err := BuildBootableImage(kernel, cmdline, noCleanup)
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