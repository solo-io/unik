package photon

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *PhotonProvider) ListVolumes() ([]*types.Volume, error) {
	return nil, errors.New("not implemented", nil)
}
