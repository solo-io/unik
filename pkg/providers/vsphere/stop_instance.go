package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxlog"
)

func (p *VsphereProvider) StopInstance(logger lxlog.Logger, id string) error {
	instance, err := p.GetInstance(logger, id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	c := p.getClient()
	err = c.PowerOffVm(logger, id)
	if err != nil {
		return lxerrors.New("failed to stop instance "+instance.Id, err)
	}
	return nil
}
