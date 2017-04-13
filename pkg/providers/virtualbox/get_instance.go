package virtualbox

import (
	"github.com/cf-unik/unik/pkg/providers/common"
	"github.com/cf-unik/unik/pkg/types"
)

func (p *VirtualboxProvider) GetInstance(nameOrIdPrefix string) (*types.Instance, error) {
	return common.GetInstance(p, nameOrIdPrefix)
}
