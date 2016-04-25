package virtualbox

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"os"
	"path/filepath"
	"time"
)

func (p *VirtualboxProvider) CreateVolume(name, imagePath string) (_ *types.Volume, err error) {
	if _, volumeErr := p.GetImage(name); volumeErr == nil {
		return nil, lxerrors.New("volume already exists", nil)
	}
	volumePath := getVolumePath(name)
	if err := os.MkdirAll(filepath.Dir(volumePath), 0777); err != nil {
		return nil, lxerrors.New("creating directory for volume file", err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(filepath.Dir(volumePath))
		}
	}()
	logrus.WithField("raw-image", imagePath).Infof("creating volume from raw image")
	if err := common.ConvertRawImage("vmdk", imagePath, volumePath); err != nil {
		return nil, lxerrors.New("converting raw image to vmdk", err)
	}

	rawImageFile, err := os.Stat(imagePath)
	if err != nil {
		return nil, lxerrors.New("statting raw image file", err)
	}
	sizeMb := rawImageFile.Size() >> 20

	volume := &types.Volume{
		Id:             name,
		Name:           name,
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
		return nil, lxerrors.New("modifying volume map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return nil, lxerrors.New("saving volume map to state", err)
	}
	return volume, nil
}
