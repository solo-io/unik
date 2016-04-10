package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
)

func (p *VsphereProvider) GetInstance(logger lxlog.Logger, nameOrIdPrefix string) (*types.Instance, error) {
	return common.GetInstance(logger, p, nameOrIdPrefix)
}
