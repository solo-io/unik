package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
)

func (p *VsphereProvider) DeleteInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	c := p.getClient()
	err = c.DestroyVm(id)
	if err != nil {
		return lxerrors.New("failed to terminate instance "+instance.Id, err)
	}
	return p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		delete(instances, instance.Id)
		return nil
	})
}
