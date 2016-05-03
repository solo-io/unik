package vsphere

import (
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *VsphereProvider) StartInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	c := p.getClient()
	err = c.PowerOnVm(instance.Name)
	if err != nil {
		return errors.New("failed to start instance "+instance.Id, err)
	}
	return nil
}
