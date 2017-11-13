package osv

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	unikos "github.com/solo-io/unik/pkg/os"
	"github.com/solo-io/unik/pkg/types"
	unikutil "github.com/solo-io/unik/pkg/util"
	"gopkg.in/yaml.v2"
)

const OSV_DEFAULT_SIZE = "1GB"

type dynamicProjectConfig struct {
	// Size is a string representing logical image size e.g. "10GB"
	Size string `yaml:"image_size"`
}

// CreateImageDynamic creates OSv image from project source directory and returns filepath of it.
func CreateImageDynamic(params types.CompileImageParams, useEc2Bootstrap bool) (string, error) {
	container := unikutil.NewContainer("compilers-osv-dynamic").
		WithVolume(params.SourcesDir+"/", "/project_directory").
		WithEnv("MAX_IMAGE_SIZE", fmt.Sprintf("%dMB", params.SizeMB))

	logrus.WithFields(logrus.Fields{
		"params": params,
	}).Debugf("running compilers-osv-dynamic container")

	if err := container.Run(); err != nil {
		return "", errors.New("failed running compilers-osv-dynamic on "+params.SourcesDir, err)
	}

	resultFile, err := ioutil.TempFile("", "osv-dynamic.qemu.")
	if err != nil {
		return "", errors.New("failed to create tmpfile for result", err)
	}
	defer func() {
		if err != nil && !params.NoCleanup {
			os.Remove(resultFile.Name())
		}
	}()

	if err := os.Rename(filepath.Join(params.SourcesDir, "boot.qcow2"), resultFile.Name()); err != nil {
		return "", errors.New("failed to rename result file", err)
	}
	return resultFile.Name(), nil
}

// readImageSizeFromManifest parses manifest.yaml and returns image size.
// Returns default image size if anything goes wrong.
func readImageSizeFromManifest(projectDir string) unikos.MegaBytes {
	config := dynamicProjectConfig{
		Size: OSV_DEFAULT_SIZE,
	}
	defaultMB, _ := unikos.ParseSize(OSV_DEFAULT_SIZE)

	data, err := ioutil.ReadFile(filepath.Join(projectDir, "manifest.yaml"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":         err,
			"defaultSize": OSV_DEFAULT_SIZE,
		}).Warning("could not find manifest.yaml. Fallback to using default unikernel size.")
		return defaultMB
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		logrus.WithFields(logrus.Fields{
			"err":         err,
			"defaultSize": OSV_DEFAULT_SIZE,
		}).Warning("failed to parse manifest.yaml. Fallback to using default unikernel size.")
		return defaultMB
	}

	sizeMB, err := unikos.ParseSize(config.Size)
	if err != nil {
		return defaultMB
	}

	return sizeMB
}
