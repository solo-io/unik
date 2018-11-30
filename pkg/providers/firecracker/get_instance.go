package firecracker

import (
	"github.com/solo-io/unik/pkg/providers/common"
	"github.com/solo-io/unik/pkg/types"
)

func (p *FirecrackerProvider) GetInstance(nameOrIdPrefix string) (*types.Instance, error) {
	return common.GetInstance(p, nameOrIdPrefix)
}
