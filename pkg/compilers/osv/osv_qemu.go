package osv

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

const OSV_QEMU_DEFAULT_MEMORY = 512
const OSV_QEMU_DEFAULT_SIZE = "10GB"

type QemuImageFinisher struct {
	ImageFinisher
}

func (b *QemuImageFinisher) FinishImage(params FinishParams) (*types.RawImage, error) {
	return &types.RawImage{
		LocalImagePath: params.CapstanImagePath,
		StageSpec: types.StageSpec{
			ImageFormat: types.ImageFormat_QCOW2,
		},
		RunSpec: types.RunSpec{
			StorageDriver:         types.StorageDriver_SATA,
			DefaultInstanceMemory: OSV_QEMU_DEFAULT_MEMORY,
			MinInstanceDiskMB:     int(readImageSizeFromManifestMB(params.CompileParams.SourcesDir)),
		},
	}, nil
}

func (b *QemuImageFinisher) UseEc2() bool {
	return false
}
