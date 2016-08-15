package xen

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *XenProvider) ListVolumes() ([]*types.Volume, error) {
	volumes := []*types.Volume{}
	for _, volume := range p.state.GetVolumes() {
		volumes = append(volumes, volume)
	}
	return volumes, nil
}
