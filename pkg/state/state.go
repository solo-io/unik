package state

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

type State interface {
	GetImages() map[string]*types.Image
	GetInstances() map[string]*types.Instance
	GetVolumes() map[string]*types.Volume
	ModifyImages(modify func(images map[string]*types.Image) error) error
	ModifyInstances(modify func(instances map[string]*types.Instance) error) error
	ModifyVolumes(modify func(volumes map[string]*types.Volume) error) error
}
