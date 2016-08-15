package xen

import "github.com/emc-advanced-dev/pkg/errors"

func (p *XenProvider) AttachVolume(id, instanceId, mntPoint string) error {
	return errors.New("not yet supportded for xen", nil)
}
