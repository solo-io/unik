package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *VsphereProvider) GetInstance(nameOrIdPrefix string) (*types.Instance, error) {
	return common.GetInstance(p, nameOrIdPrefix)
}
