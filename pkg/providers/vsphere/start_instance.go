package vsphere

import (
	"github.com/layer-x/layerx-commons/lxerrors"
)

func (p *VsphereProvider) StartInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	c := p.getClient()
	err = c.PowerOnVm(id)
	if err != nil {
		return lxerrors.New("failed to start instance "+instance.Id, err)
	}
	return nil
}
