package virtualbox

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxlog"
	"os"
)

func (p *VirtualboxProvider) DeleteVolume(logger lxlog.Logger, id string, force bool) error {
	volume, err := p.GetVolume(logger, id)
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

	p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		delete(volumes, volume.Id)
		return nil
	})

	p.state.ModifyVolumePaths(func(volumePaths map[string]string) error {
		delete(volumePaths, volume.Id)
		return nil
	})
	return nil
}
