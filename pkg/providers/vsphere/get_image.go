package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
)

func (p *VsphereProvider) GetImage(logger lxlog.Logger, nameOrIdPrefix string) (*types.Image, error) {
	return common.GetImage(logger, p, nameOrIdPrefix)
}
