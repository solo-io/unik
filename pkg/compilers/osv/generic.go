package osv

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

// OSvCompilerBase should be embedded in all OSv compilers.
type OSvCompilerBase struct {
	CompilerHelper CompilerHelper
}

// ConvertParams contains all the information needed when bootstrapping.
type ConvertParams struct {
	// CapstanImagePath points to image that was composed by Capstan
	CapstanImagePath string

	// CompileParams stores parameters that were used for composing image
	CompileParams types.CompileImageParams
}

// CompilerHelper implements conversion of Capstan result into provider-specific image.
// It should be implemented per provider. In this context converting means e.g.
// converting .qcow2 image (Capstan result) into .wmdk for VirtualBox provider.
type CompilerHelper interface {
	// Convert converts Capstan-provided image into appropriate format
	// for the provider
	Convert(params ConvertParams) (*types.RawImage, error)

	// UseEc2 tells whether or not to prepare image for EC2 (Elastic Compute Cloud)
	UseEc2() bool
}

// CompilerHelperBase should be embedded in every CompilerHelper implementation.
type CompilerHelperBase struct{}

func (b *CompilerHelperBase) UseEc2() bool {
	return false
}
