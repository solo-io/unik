package photon

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *PhotonProvider) DeleteInstance(id string, force bool) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	if instance.State == types.InstanceState_Running && !force {
		return errors.New("instance "+instance.Id+"is still running. try again with --force or power off instance first", err)
	}

	task, err := p.client.VMs.Delete(instance.Id)
	if err != nil {
		return errors.New("Delete vm", err)
	}

	task, err = p.waitForTaskSuccess(task)
	if err != nil {
		return errors.New("Delete vm", err)
	}
	err = p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		delete(instances, instance.Id)
		return nil
	})
	if err != nil {
		return errors.New("modifying image map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return errors.New("saving image map to state", err)
	}
	return nil
}
