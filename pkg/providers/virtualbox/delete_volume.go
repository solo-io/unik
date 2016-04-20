package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"os"
)

func (p *VirtualboxProvider) DeleteVolume(id string, force bool) error {
	volume, err := p.GetVolume(id)
	if err != nil {
		return lxerrors.New("retrieving volume "+id, err)
	}
	if volume.Attachment != "" {
		return lxerrors.New("volume is still attached to instance "+volume.Attachment, err)
	}
	volumePath := getVolumePath(volume.Name)
	err = os.Remove(volumePath)
	if err != nil {
		return lxerrors.New("could not delete volume at path "+volumePath, err)
	}
	err = p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		delete(volumes, volume.Id)
		return nil
	})
	if err != nil {
		return nil, lxerrors.New("deleting volume path from state", err)
	}
	err = p.state.Save()
	if err != nil {
		return lxerrors.New("saving image map to state", err)
	}

	return nil
}
