package ukvm

import (
	"github.com/cf-unik/unik/pkg/providers/common"
	"github.com/cf-unik/unik/pkg/types"
)

func (p *UkvmProvider) GetVolume(nameOrIdPrefix string) (*types.Volume, error) {
	return common.GetVolume(p, nameOrIdPrefix)
}
