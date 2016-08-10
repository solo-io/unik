package virtualbox

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
)

func (p *VirtualboxProvider) StartInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	if err := virtualboxclient.PowerOnVm(instance.Id); err != nil {
		return errors.New("failed to start instance "+instance.Id, err)
	}
	return nil
}
