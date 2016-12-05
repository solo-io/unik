package osv

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

const OSV_VMWARE_MEMORY = 512

type VmwareCompilerHelper struct {
	CompilerHelperBase
}

func (b *VmwareCompilerHelper) Convert(params ConvertParams) (*types.RawImage, error) {
	return &types.RawImage{
		LocalImagePath: params.CapstanImagePath,
		StageSpec: types.StageSpec{
			ImageFormat: types.ImageFormat_QCOW2,
		},
		RunSpec: types.RunSpec{
			DeviceMappings: []types.DeviceMapping{
				types.DeviceMapping{MountPoint: "/", DeviceName: "/dev/sda1"},
			},
			StorageDriver:         types.StorageDriver_IDE,
			VsphereNetworkType:    types.VsphereNetworkType_VMXNET3,
			DefaultInstanceMemory: OSV_VMWARE_MEMORY,
		},
	}, nil
}
