package osv

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

const OSV_VIRTUALBOX_MEMORY = 512

type VirtualboxCompilerHelper struct {
	CompilerHelperBase
}

func (b *VirtualboxCompilerHelper) Convert(params ConvertParams) (*types.RawImage, error) {
	return &types.RawImage{
		LocalImagePath: params.CapstanImagePath,
		StageSpec: types.StageSpec{
			ImageFormat: types.ImageFormat_QCOW2,
		},
		RunSpec: types.RunSpec{
			DeviceMappings: []types.DeviceMapping{
				types.DeviceMapping{MountPoint: "/", DeviceName: "/dev/sda1"},
			},
			StorageDriver:         types.StorageDriver_SATA,
			DefaultInstanceMemory: OSV_VIRTUALBOX_MEMORY,
		},
	}, nil
}
