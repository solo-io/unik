package common

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"os/exec"
)

func ConvertRawImage(imageType, inputFile, outputFile string) error {
	cmd := exec.Command("qemu-img", "convert", "-f", "raw", "-O", imageType, inputFile, outputFile)
	logrus.WithField("command", cmd.Args).Debugf("running qemu-img command")
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.New("failed converting raw image to "+imageType+": "+string(out), err)
	}
	return nil
}
