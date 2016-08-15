package xen

import "github.com/emc-advanced-dev/pkg/errors"

func (p *XenProvider) StopInstance(id string) error {
	return errors.New("Stopping xen instance is not supported", nil)
}
