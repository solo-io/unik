package vsphere

import (
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *VsphereProvider) StopInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	c := p.getClient()
	err = c.PowerOffVm(instance.Name)
	if err != nil {
		return errors.New("failed to stop instance "+instance.Id, err)
	}
	return nil
}
