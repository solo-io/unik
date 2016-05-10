package compilers

import "github.com/emc-advanced-dev/unik/pkg/types"

type Compiler interface {
	CompileRawImage(params types.CompileImageParams) (*types.RawImage, error)
}
