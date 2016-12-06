package osv

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

const OSV_VIRTUALBOX_MEMORY = 512

type VirtualboxImageFinisher struct {}

func (b *VirtualboxImageFinisher) FinishImage(params FinishParams) (*types.RawImage, error) {
	return &types.RawImage{
		LocalImagePath: params.CapstanImagePath,
		StageSpec: types.StageSpec{
			ImageFormat: types.ImageFormat_QCOW2,
		},
		RunSpec: types.RunSpec{
			DeviceMappings: []types.DeviceMapping{
				{MountPoint: "/", DeviceName: "/dev/sda1"},
			},
			StorageDriver:         types.StorageDriver_SATA,
			DefaultInstanceMemory: OSV_VIRTUALBOX_MEMORY,
		},
	}, nil
}

func (b *VirtualboxImageFinisher) UseEc2() bool {
	return false
}
