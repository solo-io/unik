package osv

import (
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

type OSvJavaCompiler struct {
	ImageFinisher ImageFinisher
}

func (r *OSvJavaCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {

	// Prepare meta/run.yaml for node runtime.
	if err := addRuntimeStanzaToMetaRun(params.SourcesDir, "java"); err != nil {
		return nil, err
	}

	// Create meta/package.yaml if not exist.
	if err := assureMetaPackage(params.SourcesDir); err != nil {
		return nil, err
	}

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

func (r *OSvJavaCompiler) Usage() *compilers.CompilerUsage {
	return &compilers.CompilerUsage{
		GeneralDescription: `
Language "java" allows you to run your Java 1.7 application.
Please provide meta/run.yaml file where you describe your project
structure. See below for more details.
`,
		PrepareApplication: "Compile your code into .class files using javac.",
		ConfigurationFiles: map[string]string{
			"/meta/run.yaml": `
config_set:
   conf1:      
      main: <name>      
      classpath:
         <list>      
      args:
         <list> 
      jvmargs:
         <list>
config_set_default: conf1
`,
			"/manifest.yaml": `
image_size: "10GB"  # logical image size
`,
		},
	}
}
