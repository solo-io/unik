package qemu

import (
	"github.com/cf-unik/unik/pkg/providers/common"
	"github.com/cf-unik/unik/pkg/types"
)

func (p *QemuProvider) GetVolume(nameOrIdPrefix string) (*types.Volume, error) {
	return common.GetVolume(p, nameOrIdPrefix)
}
