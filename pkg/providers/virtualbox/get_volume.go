package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
)

func (p *VirtualboxProvider) GetVolume(logger lxlog.Logger, nameOrIdPrefix string) (*types.Volume, error) {
	return common.GetVolume(logger, p, nameOrIdPrefix)
}
