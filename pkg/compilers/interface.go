package compilers

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"io"
)

type Compiler interface {
	CompileRawImage(sourceTar io.ReadCloser, args string, mntPoints []string) (*types.RawImage, error)
}
