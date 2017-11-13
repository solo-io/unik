package osv

import (
	"github.com/solo-io/unik/pkg/types"
)

const OSV_AWS_MEMORY = 1024

type AwsImageFinisher struct {}

func (b *AwsImageFinisher) FinishImage(params FinishParams) (*types.RawImage, error) {
	return &types.RawImage{
		LocalImagePath: params.CapstanImagePath,
		StageSpec: types.StageSpec{
			ImageFormat:           types.ImageFormat_QCOW2,
			XenVirtualizationType: types.XenVirtualizationType_HVM,
		},
		RunSpec: types.RunSpec{
			DeviceMappings: []types.DeviceMapping{
				{MountPoint: "/", DeviceName: "/dev/sda1"},
			},
			DefaultInstanceMemory: OSV_AWS_MEMORY,
		},
	}, nil
}

func (b *AwsImageFinisher) UseEc2() bool {
	return true
}
