package photon

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/types"
)

func (p *PhotonProvider) ListVolumes() ([]*types.Volume, error) {
	return nil, errors.New("not implemented", nil)
}
