package osv

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"gopkg.in/yaml.v2"
)


type javaProjectConfig struct {
	ArtifactFilename string `yaml:"artifact_filename"`
	BuildCmd string `yaml:"build_command"`
	MainClassName string `yaml:"main_class_name"`
	GroupId string `yaml:"group_id"`
	ArtifactId string `yaml:"artifact_id"`
	Version string `yaml:"version"`
}

func compileRawImage(params types.CompileImageParams, useEc2Bootstrap bool) (string, error) {
	sourcesDir := params.SourcesDir

	var config javaProjectConfig
	data, err := ioutil.ReadFile(filepath.Join(sourcesDir, "manifest.yaml"))
	if err != nil {
		return "", errors.New("failed to read manifest.yaml file", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", errors.New("failed to parse yaml manifest.yaml file", err)
	}

	container := unikutil.NewContainer("compilers-osv-java").WithVolume("/dev", "/dev").WithVolume(sourcesDir+"/", "/project_directory")
	var args []string
	if useEc2Bootstrap {
		args = append(args, "-ec2", "true")
	}

	args = append(args, "-jarName", config.ArtifactFilename)
	args = append(args, "-args", params.Args)
	args = append(args, "-groupId", config.GroupId)
	args = append(args, "-artifactId", config.ArtifactId)
	args = append(args, "-version", config.Version)
	args = append(args, "-mainClassName", config.MainClassName)
	if config.BuildCmd != "" {
		args = append(args, "-buildCmd", config.BuildCmd)
	}

	logrus.WithFields(logrus.Fields{
		"args": args,
	}).Debugf("running compilers-osv-java container")

	if err := container.Run(args...); err != nil {
		return "", errors.New("failed running compilers-osv-java on "+sourcesDir, err)
	}

	resultFile, err := ioutil.TempFile("", "osv-boot.vmdk.")
	if err != nil {
		return "", errors.New("failed to create tmpfile for result", err)
	}
	defer func() {
		if err != nil && !params.NoCleanup {
			os.Remove(resultFile.Name())
		}
	}()

	if err := os.Rename(filepath.Join(sourcesDir, "boot.qcow2"), resultFile.Name()); err != nil {
		return "", errors.New("failed to rename result file", err)
	}
	return resultFile.Name(), nil
}
