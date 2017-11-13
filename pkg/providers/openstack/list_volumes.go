package openstack

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/types"
)

func (p *OpenstackProvider) ListVolumes() ([]*types.Volume, error) {
	return nil, errors.New("not yet supportded for openstack", nil)
}
