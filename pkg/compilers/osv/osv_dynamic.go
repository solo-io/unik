package osv

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type dynamicProjectConfig struct {
	// Size is a string representing logical image size e.g. "10GB"
	Size string `yaml:"image_size"`
}

// CreateImageDynamic creates OSv image from project source directory and returns filepath of it.
func CreateImageDynamic(params types.CompileImageParams, useEc2Bootstrap bool) (string, error) {
	container := unikutil.NewContainer("compilers-osv-dynamic").
		WithVolume(params.SourcesDir+"/", "/project_directory").
		WithEnv("MAX_IMAGE_SIZE", readImageSizeFromManifest(params.SourcesDir))

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

// readImageSizeFromManifest parses manifest.yaml and returns image size string, e.g. "10GB"
// Returns default image size if anything goes wrong
func readImageSizeFromManifest(projectDir string) string {
	config := dynamicProjectConfig{
		Size: OSV_QEMU_DEFAULT_SIZE,
	}
	data, err := ioutil.ReadFile(filepath.Join(projectDir, "manifest.yaml"))
	if err != nil {
		return OSV_QEMU_DEFAULT_SIZE
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return OSV_QEMU_DEFAULT_SIZE
	}
	// append MB unit if only number was passed
	rNoUnit, _ := regexp.Compile("^[0-9]+$")
	if match := rNoUnit.FindStringSubmatch(config.Size); len(match) == 1 {
		config.Size += "MB"
	}
	return config.Size
}

// readImageSizeFromManifestMB parses manifest.yaml and returns image size in MegaBytes
func readImageSizeFromManifestMB(projectDir string) unikos.MegaBytes {
	res, _ := unikos.ParseSize(readImageSizeFromManifest(projectDir))
	return res
}
