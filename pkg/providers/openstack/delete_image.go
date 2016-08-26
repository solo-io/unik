package openstack

import (
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *OpenstackProvider) DeleteImage(id string, force bool) error {
	return errors.New("not yet supportded for openstack", nil)
}
