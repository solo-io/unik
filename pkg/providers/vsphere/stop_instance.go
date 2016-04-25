package vsphere

import (
	"github.com/layer-x/layerx-commons/lxerrors"
)

func (p *VsphereProvider) StopInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	c := p.getClient()
	err = c.PowerOffVm(instance.Id)
	if err != nil {
		return lxerrors.New("failed to stop instance "+instance.Id, err)
	}
	return nil
}
