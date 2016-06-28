package rump

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/pkg/errors"
)

func CreateImageVmware(kernel string, args string, mntPoints, bakedEnv []string, staticIpConfig string, noCleanup bool) (*types.RawImage, error) {
	return createImageVmware(kernel, args, staticIpConfig, mntPoints, bakedEnv, noCleanup, false)
}

func CreateImageVmwareAddStub(kernel string, args string, mntPoints, bakedEnv []string, staticIpConfig string, noCleanup bool) (*types.RawImage, error) {
	return createImageVmware(kernel, args, staticIpConfig, mntPoints, bakedEnv, noCleanup, true)
}

func createImageVmware(kernel, args, staticIpConfig string, mntPoints, bakedEnv []string, noCleanup, addStub bool) (*types.RawImage, error) {
	// create rump config
	var c rumpConfig
	if bakedEnv != nil {
		c.Env = make(map[string]string)
		for i, pair := range bakedEnv {
			c.Env[fmt.Sprintf("env%d", i)] = pair
		}
	}

	argv := []string{}
	if args != "" {
		argv = strings.Split(args, " ")
	}
	c = setRumpCmdLine(c, "program.bin", argv, addStub)

	res := &types.RawImage{}
	// add root -> sd0 mapping
	res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings, types.DeviceMapping{MountPoint: "/", DeviceName: "sd0"})

	bootBlk := blk{
		Source:     "dev",
		Path:       "/dev/sd0e", // no disk label on the boot partition; so partition e is used.
		FSType:     "blk",
		MountPoint: "/bootpart",
	}
	c.Blk = append(c.Blk, bootBlk)

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

	if staticIpConfig != "" {
		staticConf := strings.Split(staticIpConfig, ",")
		if len(staticConf) != 3 {
			return nil, errors.New("static ip config should be a string in the format ADDR,NETMASK,GATWAY", nil)
		}
		addr := staticConf[0]
		mask := staticConf[1]
		gw := staticConf[2]
		c.Net = &net{
			If: c.Net.If,
			Method: Static,
			Addr: addr,
			Mask: mask,
			Gatway: gw,
		}
	}

	cmdline, err := toRumpJson(c)
	if err != nil {
		return nil, err
	}

	imgFile, err := BuildBootableImage(kernel, cmdline, true, noCleanup)
	if err != nil {
		return nil, err
	}

	res.LocalImagePath = imgFile
	res.StageSpec.ImageFormat = types.ImageFormat_RAW
	res.RunSpec.StorageDriver = types.StorageDriver_SCSI
	res.RunSpec.VsphereNetworkType = types.VsphereNetworkType_E1000
	res.RunSpec.DefaultInstanceMemory = 256
	logrus.WithField("runspec", res.RunSpec).Infof("created raw vmware image")
	return res, nil
}