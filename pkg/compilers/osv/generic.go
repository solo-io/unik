package osv

import (
	"path/filepath"
	"io/ioutil"
	"os"
	"github.com/Sirupsen/logrus"
	"os/exec"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func compileRawImage(params types.CompileImageParams, useEc2Bootstrap bool) (string, error) {
	sourcesDir := params.SourcesDir
	cmd := exec.Command("docker", "run", "--rm", "--privileged",
		"-v", "/dev/:/dev/",
		"-v", sourcesDir +"/:/project_directory/",
		"projectunik/compilers-osv-java",
	)
	if useEc2Bootstrap {
		cmd.Args = append(cmd.Args, "-ec2", "true")
	}
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running compilers-osv-java container")
	unikutil.LogCommand(cmd, true)
	if err := cmd.Run(); err != nil {
		return "", errors.New("failed running compilers-osv-java on "+ sourcesDir, err)
	}

	resultFile, err := ioutil.TempFile(unikutil.UnikTmpDir(), "osv-vmdk")
	if err != nil {
		return "", errors.New("failed to create tmpfile for result", err)
	}
	defer func(){
		if err != nil {
			os.Remove(resultFile.Name())
		}
	}()

	if err := os.Rename(filepath.Join(sourcesDir, "boot.qcow2"), resultFile.Name()); err != nil {
		return "", errors.New("failed to rename result file", err)
	}
	return resultFile.Name(), nil
}