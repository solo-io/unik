package openstack

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *OpenstackProvider) GetImage(nameOrIdPrefix string) (*types.Image, error) {
	return common.GetImage(p, nameOrIdPrefix)
}
