package osv

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
)

func compileRawImage(params types.CompileImageParams, useEc2Bootstrap bool) (string, error) {
	sourcesDir := params.SourcesDir

	container := unikutil.NewContainer("compilers-osv-java").WithVolume("/dev", "/dev").WithVolume(sourcesDir+"/", "/project_directory")
	var args []string
	if useEc2Bootstrap {
		args = append(args, "-ec2", "true")
	}

	logrus.WithFields(logrus.Fields{
		"args": args,
	}).Debugf("running compilers-osv-java container")

	if err := container.Run(args...); err != nil {
		return "", errors.New("failed running compilers-osv-java on "+sourcesDir, err)
	}

	resultFile, err := ioutil.TempFile("", "TMPosv-boot.vmdk.")
	if err != nil {
		return "", errors.New("failed to create tmpfile for result", err)
	}
	defer func() {
		if err != nil {
			os.Remove(resultFile.Name())
		}
	}()

	if err := os.Rename(filepath.Join(sourcesDir, "boot.qcow2"), resultFile.Name()); err != nil {
		return "", errors.New("failed to rename result file", err)
	}
	return resultFile.Name(), nil
}
