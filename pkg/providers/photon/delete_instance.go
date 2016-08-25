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
	if instance.State == types.InstanceState_Running {
		if !force {
			return errors.New("instance "+instance.Id+"is still running. try again with --force or power off instance first", err)
		} else {
			p.StopInstance(instance.Id)
		}
	}

	task, err := p.client.VMs.Delete(instance.Id)
	if err != nil {
		return errors.New("Delete vm", err)
	}

	task, err = p.waitForTaskSuccess(task)
	if err != nil {
		return errors.New("Delete vm", err)
	}
	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		delete(instances, instance.Id)
		return nil
	}); err != nil {
		return errors.New("modifying image map in state", err)
	}
	return nil
}
