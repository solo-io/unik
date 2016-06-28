package photon

import (
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *PhotonProvider) DeleteInstance(id string, force bool) error {
	task, err := p.client.VMs.Delete(id)
	if err != nil {
		return errors.New("Delete vm", err)
	}

	task, err = p.waitForTaskSuccess(task)
	if err != nil {
		return errors.New("Delete vm", err)
	}
	return nil
}
