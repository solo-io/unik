package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
)

func (p *VsphereProvider) GetVolume(logger lxlog.Logger, nameOrIdPrefix string) (*types.Volume, error) {
	return common.GetVolume(logger, p, nameOrIdPrefix)
}
