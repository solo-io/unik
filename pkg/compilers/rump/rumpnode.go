package rump

import (
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"

	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"gopkg.in/yaml.v2"
)

// uses rump docker conter container
// the container expectes code in /opt/code and will produce program.bin in the same folder.
// we need to take the program bin and combine with json config produce an image

const (
	BootstrapTypeEC2 = "ec2"
	BootstrapTypeUDP = "udp"
)

type RumpNodeCompiler struct {
	DockerImage   string
	BootstrapType string //ec2 vs udp
	CreateImage   func(kernel, args string, mntPoints []string) (*types.RawImage, error)
}

type nodeProjectConfig struct {
	MainFile string `yaml:"main_file"`
}

func (r *RumpNodeCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {
	args := "/code/node-wrapper.js " + params.Args

	localFolder, err := ioutil.TempDir(unikutil.UnikTmpDir(), "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(localFolder)
	logrus.Debugf("extracting uploaded files to " + localFolder)
	if err := unikos.ExtractTar(params.SourceTar, localFolder); err != nil {
		return nil, err
	}

	var config nodeProjectConfig
	data, err := ioutil.ReadFile(filepath.Join(localFolder, "manifest.yaml"))
	if err != nil {
		return nil, errors.New("failed to read manifest.yaml file", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.New("failed to parse yaml manifest.yaml file", err)
	}

	if _, err := os.Stat(filepath.Join(localFolder, config.MainFile)); err != nil || config.MainFile == "" {
		return nil, errors.New("invalid main file specified", err)
	}

	logrus.Debugf("using main file %s", config.MainFile)

	env := map[string]string{
		"MAIN_FILE":      config.MainFile,
		"BOOTSTRAP_TYPE": r.BootstrapType,
	}

	if err := execContainer(r.DockerImage, nil, []string{fmt.Sprintf("%s:%s", localFolder, "/opt/code")}, false, env); err != nil {
		return nil, err
	}

	// now we should program.bin
	resultFile := path.Join(localFolder, "program.bin")

	return r.CreateImage(resultFile, args, params.MntPoints)
}
