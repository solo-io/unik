package virtualbox

import (
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
)

func (p *VirtualboxProvider) StartInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	if err := virtualboxclient.PowerOnVm(instance.Name); err != nil {
		return lxerrors.New("failed to start instance "+instance.Id, err)
	}
	return nil
}
