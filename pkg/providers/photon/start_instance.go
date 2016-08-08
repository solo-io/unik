package photon

import (
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *PhotonProvider) StartInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	task, err := p.client.VMs.Start(instance.Id)
	if err != nil {
		return errors.New("Starting vm", err)
	}

	task, err = p.waitForTaskSuccess(task)
	if err != nil {
		return errors.New("Starting vm", err)
	}
	return nil
}
