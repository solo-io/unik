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
	volumePath, ok := p.state.GetVolumePaths()[volume.Id]
	if !ok {
		return lxerrors.New("could not find path for volume "+volume.Id, nil)
	}
	err = os.Remove(volumePath)
	if err != nil {
		return lxerrors.New("could not delete volume at path "+volumePath, err)
	}

	err = p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		delete(volumes, volume.Id)
		return nil
	})

	err = p.state.ModifyVolumePaths(func(volumePaths map[string]string) error {
		delete(volumePaths, volume.Id)
		return nil
	})
	if err != nil {
		return nil, lxerrors.New("deleting volume path from state", err)
	}

	return nil
}
