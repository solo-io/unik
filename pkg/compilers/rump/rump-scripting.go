package rump

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"gopkg.in/yaml.v2"
)

// uses rump docker conter container
// the container expectes code in /opt/code and will produce program.bin in the same folder.
// we need to take the program bin and combine with json config produce an image

const (
	BootstrapTypeEC2 = "ec2"
	BootstrapTypeUDP = "udp"
)

//compiler for building images from interpreted/scripting languages (python, javascript)
type RumpScriptCompiler struct {
	RumCompilerBase

	BootstrapType string //ec2 vs udp
	RunScriptArgs string
	ScriptEnv     []string
}

type scriptProjectConfig struct {
	MainFile string `yaml:"main_file"`
}

func (r *RumpScriptCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {
	args := r.RunScriptArgs + params.Args
	sourcesDir := params.SourcesDir
	var config scriptProjectConfig
	data, err := ioutil.ReadFile(filepath.Join(sourcesDir, "manifest.yaml"))
	if err != nil {
		return nil, errors.New("failed to read manifest.yaml file", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.New("failed to parse yaml manifest.yaml file", err)
	}

	if _, err := os.Stat(filepath.Join(sourcesDir, config.MainFile)); err != nil || config.MainFile == "" {
		return nil, errors.New("invalid main file specified", err)
	}

	logrus.Debugf("using main file %s", config.MainFile)

	containerEnv := []string{
		fmt.Sprintf("MAIN_FILE=%s", config.MainFile),
		fmt.Sprintf("BOOTSTRAP_TYPE=%s", r.BootstrapType),
	}

	if err := r.runContainer(sourcesDir, containerEnv); err != nil {
		return nil, err
	}

	// now we should program.bin
	resultFile := path.Join(sourcesDir, "program.bin")

	return r.CreateImage(resultFile, args, params.MntPoints, r.ScriptEnv, params.StaticIpConfig, params.NoCleanup)
}
