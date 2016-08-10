package osv

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

type OSvCompilerBase struct {
	CreateImage func(params types.CompileImageParams, useEc2Bootstrap bool) (string, error)
}
