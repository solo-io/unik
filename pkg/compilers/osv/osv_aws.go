package osv

import (
	"io"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/pkg/errors"
)

type OsvAwsCompiler struct {
	ExtraConfig types.ExtraConfig
}

func (osvCompiler *OsvAwsCompiler) CompileRawImage(sourceTar io.ReadCloser, args string, mntPoints []string) (_ *types.RawImage, err error) {
	resultFile, err := compileRawImage(sourceTar, args, mntPoints)
	if err != nil {
		return nil, errors.New("failed to compile raw osv image", err)
	}
	return &types.RawImage{
		LocalImagePath: resultFile,
		ExtraConfig: 	osvCompiler.ExtraConfig,
		DeviceMappings: []types.DeviceMapping{
			types.DeviceMapping{MountPoint: "/", DeviceName: "/dev/sda1"},
		},
	}, nil
}