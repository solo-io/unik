package osv

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

type OsvVmwareCompiler struct{}

const OSV_VMWARE_MEMORY = 512

func (osvCompiler *OsvVmwareCompiler) CompileRawImage(params types.CompileImageParams) (_ *types.RawImage, err error) {
	resultFile, err := compileRawImage(params, false)
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
			StorageDriver:         types.StorageDriver_IDE,
			VsphereNetworkType:    types.VsphereNetworkType_VMXNET3,
			DefaultInstanceMemory: OSV_VMWARE_MEMORY,
		},
	}, nil
}
