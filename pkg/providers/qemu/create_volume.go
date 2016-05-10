package qemu

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *QemuProvider) CreateVolume(params types.CreateVolumeParams) (*types.Volume, error) {
	return nil, errors.New("not supported", nil)
}
