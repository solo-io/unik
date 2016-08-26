package openstack

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *OpenstackProvider) CreateVolume(params types.CreateVolumeParams) (_ *types.Volume, err error) {
	return nil, errors.New("not yet supportded for openstack", nil)
}
