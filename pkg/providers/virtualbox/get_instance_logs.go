package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxlog"
)

func (p *VirtualboxProvider) GetInstanceLogs(logger lxlog.Logger, id string) (string, error) {
	instance, err := p.GetInstance(logger, id)
	if err != nil {
		return "", lxerrors.New("retrieving instance "+id, err)
	}
	return common.GetInstanceLogs(logger, instance)
}
