package openstack

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *OpenstackProvider) GetVolume(nameOrIdPrefix string) (*types.Volume, error) {
	return nil, errors.New("not yet supportded for openstack", nil)
}
