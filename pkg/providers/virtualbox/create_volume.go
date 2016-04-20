package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/layer-x/layerx-commons/lxerrors"
	"os"
	"time"
)

func (p *VirtualboxProvider) CreateVolume(name, imagePath string) (*types.Volume, error) {
	volumePath := getVolumePath(name)
	logrus.WithField("raw-image", imagePath).Infof("creating volume from raw image")
	if err := common.ConvertRawImage("vmdk", volumePath); err != nil {
		return nil, lxerrors.New("converting raw image to vmdk", err)
	}

	rawImageFile, err := os.Stat(imagePath)
	if err != nil {
		return nil, lxerrors.New("statting raw image file", err)
	}
	sizeMb := rawImageFile.Size() >> 20

	volume := &types.Volume{
		Id: name,
		Name: name,
		SizeMb: sizeMb,
		Attachment: "",
		Infrastructure: types.Infrastructure_VIRTUALBOX,
		Created: time.Now(),
	}

	err = p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error{
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
	return nil
}