package rump

import (
	"fmt"
	"strings"

	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/pkg/errors"
)

func CreateImageAws(kernel, args string, mntPoints, bakedEnv []string, staticIpConfig string, noCleanup bool) (*types.RawImage, error) {
	return createImageAws(kernel, args, staticIpConfig, mntPoints, bakedEnv, noCleanup, false)
}

func CreateImageAwsAddStub(kernel, args string, mntPoints, bakedEnv []string, staticIpConfig string, noCleanup bool) (*types.RawImage, error) {
	return createImageAws(kernel, args, staticIpConfig, mntPoints, bakedEnv, noCleanup, true)
}

func createImageAws(kernel, args, staticIpConfig string, mntPoints, bakedEnv []string, noCleanup, addStub bool) (*types.RawImage, error) {
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
	volIndex := 0
	// add root -> sda1 mapping
	res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings, types.DeviceMapping{MountPoint: "/", DeviceName: "/dev/sda1"})

	bootBlk := blk{
		Source:     "etfs",
		Path:       "sda1",
		FSType:     "blk",
		MountPoint: "/bootpart",
	}
	c.Blk = append(c.Blk, bootBlk)

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
	imgFile, err := BuildBootableImage(kernel, cmdline, false, noCleanup)
	if err != nil {
		return nil, err
	}

	res.LocalImagePath = imgFile
	res.StageSpec = types.StageSpec{
		ImageFormat:           types.ImageFormat_RAW,
		XenVirtualizationType: types.XenVirtualizationType_Paravirtual,
	}
	res.RunSpec.DefaultInstanceMemory = 256

	return res, nil
}
