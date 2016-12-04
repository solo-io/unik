package osv

import (
	"github.com/emc-advanced-dev/unik/pkg/compilers/osv/osvbootstrap"
)

// OSvCompilerBase should be embedded in all OSv compilers.
type OSvCompilerBase struct {
	Bootstrapper osvbootstrap.Bootstrapper
}
