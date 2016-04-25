package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/layer-x/layerx-commons/lxerrors"
)

func (p *VirtualboxProvider) StartInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	if err := virtualboxclient.PowerOnVm(instance.Id); err != nil {
		return lxerrors.New("failed to start instance "+instance.Id, err)
	}
	return nil
}
