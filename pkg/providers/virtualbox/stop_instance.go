package virtualbox

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
)

func (p *VirtualboxProvider) StopInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	if err := virtualboxclient.PowerOffVm(instance.Id); err != nil {
		return errors.New("failed to stop instance "+instance.Id, err)
	}
	return nil
}
