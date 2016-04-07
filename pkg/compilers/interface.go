package compilers

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"mime/multipart"
)

type Compiler interface {
	CompileRawImage(sourceTar multipart.File, tarHeader *multipart.FileHeader, mntPoints []string) (*types.RawImage, error)
}
