package compilers

import (
	"mime/multipart"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

type Compiler interface {
	CompileBootVolume(sourceTar multipart.File, tarHeader *multipart.FileHeader, mntPoints []string) (*types.BootVolume, error)
}