// Package osvbootstrap provides implementation of
// osv.Bootstrapper interface for each provider.
package osvbootstrap

import "github.com/emc-advanced-dev/unik/pkg/types"

// BootstrapParams contains all the information needed when bootstrapping.
type BootstrapParams struct {
	// CapstanImagePath points to image that was composed by Capstan
	CapstanImagePath string

	// CompileParams stores parameters that were used for composing image
	CompileParams types.CompileImageParams
}

// Bootstrapper implements conversion of Capstan result into provider-like image.
// It should be implemented per provider. In this context Bootstrapping means e.g.
// converting .qcow2 image into .wmdk for VirtualBox.
type Bootstrapper interface {
	// Bootstrap converts Capstan-provided image into appropriate format
	// for the provider
	Bootstrap(params BootstrapParams) (*types.RawImage, error)

	// UseEc2 tells whether or not to prepare image for EC2 (Elastic Compute Cloud)
	UseEc2() bool
}
