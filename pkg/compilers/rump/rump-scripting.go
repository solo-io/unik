package rump

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/compilers"
	"github.com/solo-io/unik/pkg/types"
	"gopkg.in/yaml.v2"
)

// uses rump docker conter container
// the container expectes code in /opt/code and will produce program.bin in the same folder.
// we need to take the program bin and combine with json config produce an image

const (
	BootstrapTypeEC2    = "ec2"
	BootstrapTypeUDP    = "udp"
	BootstrapTypeGCLOUD = "gcloud"
	BootstrapTypeNoStub = "nostub"
)

//compiler for building images from interpreted/scripting languages (python, javascript)
type RumpScriptCompiler struct {
	RumCompilerBase

	BootstrapType string //ec2 vs udp
	RunScriptArgs string
	ScriptEnv     []string
}

type scriptProjectConfig struct {
	MainFile    string `yaml:"main_file"`
	RuntimeArgs string `yaml:"runtime_args"`
}

func (r *RumpScriptCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {
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

	resultFile := path.Join(sourcesDir, "program.bin")

	//build args string
	args := r.RunScriptArgs
	if config.RuntimeArgs != "" {
		args = config.RuntimeArgs + " " + args
	}
	if params.Args != "" {
		args = args + " " + params.Args
	}

	return r.CreateImage(resultFile, args, params.MntPoints, append(r.ScriptEnv, fmt.Sprintf("MAIN_FILE=%s", config.MainFile), fmt.Sprintf("BOOTSTRAP_TYPE=%s", r.BootstrapType)), params.NoCleanup)
}

func (r *RumpScriptCompiler) Usage() *compilers.CompilerUsage {
	return nil
}

func NewRumpPythonCompiler(dockerImage string, createImage func(kernel, args string, mntPoints, bakedEnv []string, noCleanup bool) (*types.RawImage, error), bootStrapType string) *RumpScriptCompiler {
	return &RumpScriptCompiler{
		RumCompilerBase: RumCompilerBase{
			DockerImage: dockerImage,
			CreateImage: createImage,
		},
		BootstrapType: bootStrapType,
		RunScriptArgs: "/bootpart/python-wrapper.py",
		ScriptEnv: []string{
			"PYTHONHOME=/bootpart/python",
			"PYTHONPATH=/bootpart/lib/python3.5/site-packages/:/bootpart/bin/",
		},
	}
}

func NewRumpJavaCompiler(dockerImage string, createImage func(kernel, args string, mntPoints, bakedEnv []string, noCleanup bool) (*types.RawImage, error), bootStrapType string) *RumpScriptCompiler {
	return &RumpScriptCompiler{
		RumCompilerBase: RumCompilerBase{
			DockerImage: dockerImage,
			CreateImage: createImage,
		},
		BootstrapType: bootStrapType,
		RunScriptArgs: "-jar /bootpart/program.jar",
		ScriptEnv: []string{
			"CLASSPATH=/bootpart/jetty:/bootpart/jdk/jre/lib",
			"JAVA_HOME=/bootpart/jdk/",
		},
	}
}
