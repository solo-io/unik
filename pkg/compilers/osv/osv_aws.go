package osv

import (
	"io"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/pkg/errors"
)

type OsvAwsCompiler struct {
}

const OSV_AWS_MEMORY = 1024

func (osvCompiler *OsvAwsCompiler) CompileRawImage(sourceTar io.ReadCloser, args string, mntPoints []string) (_ *types.RawImage, err error) {
	resultFile, err := compileRawImage(sourceTar, args, mntPoints, true)
	if err != nil {
		return nil, errors.New("failed to compile raw osv image", err)
	}
	return &types.RawImage{
		LocalImagePath: resultFile,
		StageSpec: types.StageSpec{
			ImageFormat: types.ImageFormat_QCOW2,
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