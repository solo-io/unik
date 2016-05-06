package osv

import (
	"io"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/pkg/errors"
)

type OsvVmwareCompiler struct {}

func (osvCompiler *OsvVmwareCompiler) CompileRawImage(sourceTar io.ReadCloser, args string, mntPoints []string) (_ *types.RawImage, err error) {
	resultFile, err := compileRawImage(sourceTar, args, mntPoints, false)
	if err != nil {
		return nil, errors.New("failed to compile raw osv image", err)
	}
	return &types.RawImage{
		LocalImagePath: resultFile,
		StageSpec: types.StageSpec{
			ImageFormat: types.ImageFormat_QCOW2,
		},
		RunSpec: types.RunSpec{
			DeviceMappings: []types.DeviceMapping{
				types.DeviceMapping{MountPoint: "/", DeviceName: "/dev/sda1"},
			},
			StorageDriver: types.StorageDriver_IDE,
			VsphereNetworkType: types.VsphereNetworkType_VMXNET3,
		},
	}, nil
}