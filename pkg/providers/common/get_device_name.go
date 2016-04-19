package common

import (
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func GetDeviceNameForMnt(image *types.Image, mntPoint string) (string, error) {
	for _, mapping := range image.DeviceMappings {
		if mntPoint == mapping.MountPoint {
			return mapping.DeviceName, nil
		}
	}
	return "", lxerrors.New("no mapping found on image "+image.Id+" for mount point "+mntPoint, nil)
}