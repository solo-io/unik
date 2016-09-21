package openstack

import "github.com/emc-advanced-dev/pkg/errors"

func (p *OpenstackProvider) AttachVolume(id, instanceId, mntPoint string) error {
	return errors.New("not yet supportded for openstack", nil)
}
