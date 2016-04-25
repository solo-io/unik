package common

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/providers"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
)

func VerifyMntsInput(p providers.Provider, image *types.Image, mntPointsToVolumeIds map[string]string) error {
	for _, deviceMapping := range image.DeviceMappings {
		if deviceMapping.MountPoint == "/" {
			//ignore boot mount point
			continue
		}
		_, ok := mntPointsToVolumeIds[deviceMapping.MountPoint]
		if !ok {
			logrus.WithFields(logrus.Fields{"required-device-mappings": image.DeviceMappings}).Errorf("requied mount point missing: %s", deviceMapping.MountPoint)
			return lxerrors.New("required mount point missing from input", nil)
		}
	}
	for mntPoint, volumeId := range mntPointsToVolumeIds {
		mntPointExists := false
		for _, deviceMapping := range image.DeviceMappings {
			if deviceMapping.MountPoint == mntPoint {
				mntPointExists = true
				break
			}
		}
		if !mntPointExists {
			return lxerrors.New("mount point "+mntPoint+" does not exist for image "+image.Id, nil)
		}
		_, err := p.GetVolume(volumeId)
		if err != nil {
			return lxerrors.New("could not find volume "+volumeId, err)
		}
	}
	return nil
}
