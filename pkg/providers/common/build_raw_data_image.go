package common

import (
	"github.com/layer-x/layerx-commons/lxlog"
	"os/exec"
	"github.com/layer-x/layerx-commons/lxerrors"
	"path/filepath"
)

func BuildRawDataImage(logger lxlog.Logger, dataFolderPath string, size int) (string, error) {
	var cmd *exec.Cmd
	if size > 0 {
		cmd = exec.Command("docker", "run", "--rm", "--privileged",
			"-v", "/dev/:/dev/",
			"-v", filepath.Dir(dataFolderPath)+":/opt/code/",
			"image-creator",
			"-v", filepath.Base(dataFolderPath), size,
		)
	}
	logger.WithFields(lxlog.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
	logger.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc vm.destroy " + vmName, err)
	}
	return nil
}