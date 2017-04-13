package photon

import (
	"github.com/cf-unik/unik/pkg/providers/common"
	"github.com/cf-unik/unik/pkg/types"
)

func (p *PhotonProvider) GetInstance(nameOrIdPrefix string) (*types.Instance, error) {
	return common.GetInstance(p, nameOrIdPrefix)
}
