package osv

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type OSvJavaCompiler struct {
	OSvCompilerBase
}

// javaProjectConfig defines available inputs
type javaProjectConfig struct {
	MainFile    string `yaml:"main_file"`
	RuntimeArgs string `yaml:"runtime_args"`
	BuildCmd    string `yaml:"build_command"`
}

func (r *OSvJavaCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {
	sourcesDir := params.SourcesDir

	var config javaProjectConfig
	data, err := ioutil.ReadFile(filepath.Join(sourcesDir, "manifest.yaml"))
	if err != nil {
		return nil, errors.New("failed to read manifest.yaml file", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.New("failed to parse yaml manifest.yaml file", err)
	}

	container := unikutil.NewContainer("compilers-osv-java").WithVolume("/dev", "/dev").WithVolume(sourcesDir+"/", "/project_directory")
	var args []string
	if r.CompilerHelper.UseEc2() {
		args = append(args, "-ec2")
	}

	args = append(args, "-main_file", config.MainFile)
	args = append(args, "-args", params.Args)
	if config.BuildCmd != "" {
		args = append(args, "-buildCmd", config.BuildCmd)
	}
	if len(config.RuntimeArgs) > 0 {
		args = append(args, "-runtime", config.RuntimeArgs)
	}

	logrus.WithFields(logrus.Fields{
		"args": args,
	}).Debugf("running compilers-osv-java container")

	if err := container.Run(args...); err != nil {
		return nil, errors.New("failed running compilers-osv-java on "+sourcesDir, err)
	}

	// And finally bootstrap.
	convertParams := ConvertParams{
		CompileParams:    params,
		CapstanImagePath: filepath.Join(sourcesDir, "boot.qcow2"),
	}
	return r.CompilerHelper.Convert(convertParams)
}
