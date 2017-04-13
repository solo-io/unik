package ukvm

import (
	"github.com/cf-unik/unik/pkg/types"
)

func (p *UkvmProvider) ListVolumes() ([]*types.Volume, error) {
	volumes := []*types.Volume{}
	for _, volume := range p.state.GetVolumes() {
		volumes = append(volumes, volume)
	}
	return volumes, nil
}
