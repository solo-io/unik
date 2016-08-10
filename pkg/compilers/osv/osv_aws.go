package osv

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

type OsvAwsCompiler struct {
}

const OSV_AWS_MEMORY = 1024

func (osvCompiler *OsvAwsCompiler) CompileRawImage(params types.CompileImageParams) (_ *types.RawImage, err error) {
	resultFile, err := compileRawImage(params, true)
	if err != nil {
		return nil, errors.New("failed to compile raw osv image", err)
	}
	return &types.RawImage{
		LocalImagePath: resultFile,
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
