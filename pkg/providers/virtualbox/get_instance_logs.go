package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *VirtualboxProvider) GetInstanceLogs(id string) (string, error) {
	instance, err := p.GetInstance(id)
	if err != nil {
		return "", errors.New("retrieving instance "+id, err)
	}
	return common.GetInstanceLogs(instance)
}
