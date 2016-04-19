package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
)

func (p *VirtualboxProvider) DeleteInstance(id string) error {
	//TO MAKE SURE WE DONT DELETE VOLUME BEFORE INSTANCE DDELETE, POWER DOWN AND DETACH FIRST! :)
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
