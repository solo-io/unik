package common

import (
	"os/exec"
	"github.com/layer-x/layerx-commons/lxerrors"
	"path/filepath"
	"github.com/Sirupsen/logrus"
	uniklog "github.com/emc-advanced-dev/unik/pkg/util/log"
	"fmt"
)

func BuildRawDataImage(dataFolderPath string, size int) (string, error) {
	var cmd *exec.Cmd
	if size > 0 {
		cmd = exec.Command("docker", "run", "--rm", "--privileged",
			"-v", "/dev/:/dev/",
			"-v", filepath.Dir(dataFolderPath)+":/opt/code/",
			"image-creator",
			"-v", filepath.Base(dataFolderPath), fmt.Sprintf("%v", size),
		)
	}
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return "", lxerrors.New("failed running image-creator on " + dataFolderPath, err)
	}
	return "", nil
}