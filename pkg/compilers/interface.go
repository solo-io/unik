package compilers

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

type Compiler interface {
	CompileRawImage(params types.CompileImageParams) (*types.RawImage, error)

	// Usage describes how to prepare project to run it with UniK
	// The returned text should describe what configuration files to
	// prepare and how.
	Usage() string
}
