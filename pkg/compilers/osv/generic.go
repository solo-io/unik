package osv

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

// ConvertParams contains all the information needed when bootstrapping.
type FinishParams struct {
	// CapstanImagePath points to image that was composed by Capstan
	CapstanImagePath string

	// CompileParams stores parameters that were used for composing image
	CompileParams types.CompileImageParams
}

// ImageFinisher implements conversion of Capstan result into provider-specific image.
// It should be implemented per provider. In this context converting means e.g.
// converting .qcow2 image (Capstan result) into .wmdk for VirtualBox provider.
type ImageFinisher interface {
	// Convert converts Capstan-provided image into appropriate format
	// for the provider
	FinishImage(params FinishParams) (*types.RawImage, error)

	// UseEc2 tells whether or not to prepare image for EC2 (Elastic Compute Cloud)
	UseEc2() bool
}
