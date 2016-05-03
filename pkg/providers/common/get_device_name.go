package common

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/pkg/errors"
)

func GetDeviceNameForMnt(image *types.Image, mntPoint string) (string, error) {
	for _, mapping := range image.DeviceMappings {
		if mntPoint == mapping.MountPoint {
			return mapping.DeviceName, nil
		}
	}
	return "", errors.New("no mapping found on image "+image.Id+" for mount point "+mntPoint, nil)
}

func GetControllerPortForMnt(image *types.Image, mntPoint string) (int, error) {
	for controllerPort, mapping := range image.DeviceMappings {
		if mntPoint == mapping.MountPoint {
			return controllerPort, nil
		}
	}
	return -1, errors.New("no mapping found on image "+image.Id+" for mount point "+mntPoint, nil)
}
