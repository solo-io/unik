package vsphere

import (
	"github.com/cf-unik/unik/pkg/providers/common"
	"github.com/cf-unik/unik/pkg/types"
)

func (p *VsphereProvider) GetImage(nameOrIdPrefix string) (*types.Image, error) {
	return common.GetImage(p, nameOrIdPrefix)
}
