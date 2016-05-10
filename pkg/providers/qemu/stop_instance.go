package qemu

import (
	"strconv"
	"syscall"

	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *QemuProvider) StopInstance(id string) error {

	// TODO:
	// kill qemu
	pid, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("failed parsing qemu id", err)
	}

	return syscall.Kill(pid, syscall.SIGKILL)
}
