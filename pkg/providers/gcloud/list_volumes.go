package gcloud

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/types"
)

func (p *GcloudProvider) ListVolumes() ([]*types.Volume, error) {
	return nil, errors.New("not yet implemented", nil)
}
