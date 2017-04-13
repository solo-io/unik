package rump

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/compilers"
	"github.com/cf-unik/unik/pkg/types"
	"gopkg.in/yaml.v2"
)

//compiler for building images from interpreted/scripting languages (python, javascript)
type RumpCCompiler struct {
	RumCompilerBase
}

type cProjectConfig struct {
	BinaryName string `yaml:"binary_name"`
}

func (r *RumpCCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {
	sourcesDir := params.SourcesDir
	var config cProjectConfig
	data, err := ioutil.ReadFile(filepath.Join(sourcesDir, "manifest.yaml"))
	if err != nil {
		return nil, errors.New("failed to read manifest.yaml file", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.New("failed to parse yaml manifest.yaml file", err)
	}

	containerEnv := []string{
		fmt.Sprintf("BINARY_NAME=%s", config.BinaryName),
	}

	if err := r.runContainer(sourcesDir, containerEnv); err != nil {
		return nil, err
	}

	resultFile := path.Join(sourcesDir, "program.bin")

	return r.CreateImage(resultFile, params.Args, params.MntPoints, nil, params.NoCleanup)
}

func (r *RumpCCompiler) Usage() *compilers.CompilerUsage {
	return nil
}

func NewRumpCCompiler(dockerImage string, createImage func(kernel, args string, mntPoints, bakedEnv []string, noCleanup bool) (*types.RawImage, error)) *RumpCCompiler {
	return &RumpCCompiler{
		RumCompilerBase: RumCompilerBase{
			DockerImage: dockerImage,
			CreateImage: createImage,
		},
	}
}
