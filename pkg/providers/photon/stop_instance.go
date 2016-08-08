package photon

import "github.com/emc-advanced-dev/pkg/errors"

func (p *PhotonProvider) StopInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	task, err := p.client.VMs.Stop(instance.Id)
	if err != nil {
		return errors.New("Stopping vm", err)
	}

	task, err = p.waitForTaskSuccess(task)
	if err != nil {
		return errors.New("Stopping vm", err)
	}
	return nil
}
