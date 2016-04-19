package common

import (
	"os/exec"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/Sirupsen/logrus"
)

func ConvertRawImage(imageType, outputFile string) (error) {
	cmd := exec.Command("qemu-img", "convert", "-f", "raw", "-O", imageType, outputFile)
	logrus.WithField("command", cmd.Args).Debugf("running qemu-img command")
	if out, err := cmd.CombinedOutput(); err != nil {
		return lxerrors.New("failed converting raw image to "+imageType+": "+string(out), err)
	}
	return nil
}