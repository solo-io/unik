package ukvm

import (
	"github.com/cf-unik/unik/pkg/providers/common"
	"github.com/cf-unik/unik/pkg/types"
)

func (p *UkvmProvider) GetImage(nameOrIdPrefix string) (*types.Image, error) {
	return common.GetImage(p, nameOrIdPrefix)
}
