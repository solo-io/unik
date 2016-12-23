package osv

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

const OSV_QEMU_DEFAULT_MEMORY = 512
const OSV_QEMU_DEFAULT_SIZE = "10GB"

type OsvQemuCompiler struct {
	OSvCompilerBase
}

func (osvCompiler *OsvQemuCompiler) CompileRawImage(params types.CompileImageParams) (_ *types.RawImage, err error) {
	resultFile, err := osvCompiler.CreateImage(params, false)
	if err != nil {
		return nil, errors.New("failed to compile raw OSv dynamic image", err)
	}
	return &types.RawImage{
		LocalImagePath: resultFile,
		StageSpec: types.StageSpec{
			ImageFormat: types.ImageFormat_QCOW2,
		},
		RunSpec: types.RunSpec{
			StorageDriver:         types.StorageDriver_SATA,
			DefaultInstanceMemory: OSV_QEMU_DEFAULT_MEMORY,
			MinInstanceDiskMB:     int(readImageSizeFromManifestMB(params.SourcesDir)),
		},
	}, nil
}
