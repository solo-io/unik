package osv

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

const OSV_AWS_MEMORY = 1024

type AwsCompilerHelper struct {
	CompilerHelperBase
}

func (b *AwsCompilerHelper) Convert(params ConvertParams) (*types.RawImage, error) {
	return &types.RawImage{
		LocalImagePath: params.CapstanImagePath,
		StageSpec: types.StageSpec{
			ImageFormat:           types.ImageFormat_QCOW2,
			XenVirtualizationType: types.XenVirtualizationType_HVM,
		},
		RunSpec: types.RunSpec{
			DeviceMappings: []types.DeviceMapping{
				types.DeviceMapping{MountPoint: "/", DeviceName: "/dev/sda1"},
			},
			DefaultInstanceMemory: OSV_AWS_MEMORY,
		},
	}, nil
}

func (b *AwsCompilerHelper) UseEc2() bool {
	return true
}
