package virtualbox

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"os"
	"path/filepath"
	"time"
)

func (p *VirtualboxProvider) CreateVolume(params types.CreateVolumeParams) (_ *types.Volume, err error) {
	if _, volumeErr := p.GetImage(params.Name); volumeErr == nil {
		return nil, errors.New("volume already exists", nil)
	}
	volumePath := getVolumePath(params.Name)
	if err := os.MkdirAll(filepath.Dir(volumePath), 0755); err != nil {
		return nil, errors.New("creating directory for volume file", err)
	}
	defer func() {
		if params.NoCleanup {
			logrus.Warnf("because --no-cleanup flag was provided, not cleaning up failed volume %s at %s", params.Name, volumePath)
			return
		}
		if err != nil {
			os.RemoveAll(filepath.Dir(volumePath))
		}
	}()
	logrus.WithField("raw-image", params.ImagePath).Infof("creating volume from raw image")
	if err := common.ConvertRawImage(types.ImageFormat_RAW, types.ImageFormat_VMDK, params.ImagePath, volumePath); err != nil {
		return nil, errors.New("converting raw image to vmdk", err)
	}

	rawImageFile, err := os.Stat(params.ImagePath)
	if err != nil {
		return nil, errors.New("statting raw image file", err)
	}
	sizeMb := rawImageFile.Size() >> 20

	volume := &types.Volume{
		Id:             params.Name,
		Name:           params.Name,
		SizeMb:         sizeMb,
		Attachment:     "",
		Infrastructure: types.Infrastructure_VIRTUALBOX,
		Created:        time.Now(),
	}

	err = p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		volumes[volume.Id] = volume
		return nil
	})
	if err != nil {
		return nil, errors.New("modifying volume map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return nil, errors.New("saving volume map to state", err)
	}
	return volume, nil
}
