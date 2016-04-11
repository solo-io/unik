package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
)

func (p *VirtualboxProvider) GetInstance(logger lxlog.Logger, nameOrIdPrefix string) (*types.Instance, error) {
	return common.GetInstance(logger, p, nameOrIdPrefix)
}
