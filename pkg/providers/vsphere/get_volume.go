package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *VsphereProvider) GetVolume(nameOrIdPrefix string) (*types.Volume, error) {
	return common.GetVolume(p, nameOrIdPrefix)
}
