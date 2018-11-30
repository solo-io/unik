package firecracker

import (
	"github.com/solo-io/unik/pkg/providers/common"
	"github.com/solo-io/unik/pkg/types"
)

func (p *FirecrackerProvider) GetVolume(nameOrIdPrefix string) (*types.Volume, error) {
	return common.GetVolume(p, nameOrIdPrefix)
}
