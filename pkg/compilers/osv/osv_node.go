package osv

import (
	"github.com/cf-unik/unik/pkg/compilers"
	"github.com/cf-unik/unik/pkg/types"
)

type OSvNodeCompiler struct {
	ImageFinisher ImageFinisher
}

func (r *OSvNodeCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {

	// Prepare meta/run.yaml for node runtime.
	if err := addRuntimeStanzaToMetaRun(params.SourcesDir, "node"); err != nil {
		return nil, err
	}

	// Create meta/package.yaml if not exist.
	if err := assureMetaPackage(params.SourcesDir); err != nil {
		return nil, err
	}

	// Parse image size from manifest.yaml.
	params.SizeMB = int(readImageSizeFromManifest(params.SourcesDir))

	// Compose image inside Docker container.
	imagePath, err := CreateImageDynamic(params, r.ImageFinisher.UseEc2())
	if err != nil {
		return nil, err
	}

	// And finalize it.
	convertParams := FinishParams{
		CompileParams:    params,
		CapstanImagePath: imagePath,
	}
	return r.ImageFinisher.FinishImage(convertParams)
}

func (r *OSvNodeCompiler) Usage() *compilers.CompilerUsage {
	return &compilers.CompilerUsage{
		PrepareApplication: "Install all libraries using `npm install`.",
		ConfigurationFiles: map[string]string{
			"/meta/run.yaml": `
config_set:
   conf1:
      main: <relative-path-to-your-entrypoint>   
config_set_default: conf1
`,
			"/manifest.yaml": `
image_size: "10GB"  # logical image size
`,
		},
	}
}
