package compilers

import (
	"io"

	"github.com/emc-advanced-dev/unik/pkg/types"
)

type Compiler interface {
	CompileRawImage(sourceTar io.ReadCloser, args string, mntPoints []string) (*types.RawImage, error)
}
