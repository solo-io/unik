package osv

import (
	"fmt"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

type OSvNativeCompiler struct {
	ImageFinisher ImageFinisher
}

func (r *OSvNativeCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {

	// Prepare meta/run.yaml for node runtime.
	if err := addRuntimeStanzaToMetaRun(params.SourcesDir, "native"); err != nil {
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

func (r *OSvNativeCompiler) Usage() *compilers.CompilerUsage {
	return &compilers.CompilerUsage{
		GeneralDescription: `
Language "native" allows you to either:
a) run precompiled package from the remote repository
b) run your own binary code
c) combination of (a) and (b)
In case you intend to use packages from the remote repository, please provide
meta/run.yaml where you list desired packages.
In any case (i.e. (a), (b), (c)) you need to provide meta/run.yaml. See below for
more details.
`,
		PrepareApplication: `
(this is only needed if you want to run your own C/C++ application)
Compile your application into relocatable shared-object (a file normally
given a ".so" extension) that is PIC (position independent code).
`,
		ConfigurationFiles: map[string]string{
			"/meta/run.yaml": `
config_set:
   conf1:
      bootcmd: <boot-command-that-starts-application>    
config_set_default: conf1
`,
			"/meta/package.yaml": `
title: <your-unikernel-title>
name: <your-unikernel-name>
author: <your-name>
require:
  - <first-required-package-title>
  - <second-required-package-title>
  # ...
`,
			"/manifest.yaml": `
image_size: "10GB"  # logical image size
`,
		},
		Other: fmt.Sprintf(`
Below please find a list of packages in remote repository:
%s
`, listOfPackages()),
	}
}

func listOfPackages() string {
	return `
eu.mikelangelo-project.app.hadoop-hdfs
eu.mikelangelo-project.app.mysql-5.6.21
eu.mikelangelo-project.erlang
eu.mikelangelo-project.ompi
eu.mikelangelo-project.openfoam.core
eu.mikelangelo-project.openfoam.pimplefoam
eu.mikelangelo-project.openfoam.pisofoam
eu.mikelangelo-project.openfoam.poroussimplefoam
eu.mikelangelo-project.openfoam.potentialfoam
eu.mikelangelo-project.openfoam.rhoporoussimplefoam
eu.mikelangelo-project.openfoam.rhosimplefoam
eu.mikelangelo-project.openfoam.simplefoam
eu.mikelangelo-project.osv.cli
eu.mikelangelo-project.osv.cloud-init
eu.mikelangelo-project.osv.httpserver
eu.mikelangelo-project.osv.nfs
`
}
