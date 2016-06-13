// +build cgo

package qemu

import (
	"os"
	"strings"
	"syscall"

	"github.com/digitalocean/go-ps"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *QemuProvider) StopInstance(id string) error {

	instance, err := p.GetInstance(id)
	if err != nil {
		return err
	}
	image, err := p.GetImage(instance.ImageId)
	if err != nil {
		return err
	}

	proc, err := getOurQemu(image)
	if err != nil {
		return err
	}

	// kill qemu
	return syscall.Kill(proc.Pid(), syscall.SIGKILL)
}

func getOurQemu(image *types.Image) (ps.Process, error) {

	procs, err := ps.Processes()
	if err != nil {
		return nil, err
	}

	instanceArg := getKernelPath(image.Name)
	for _, proc := range procs {
		if !strings.Contains(proc.Executable(), "qemu") {
			continue
		}

		// qemu must belong either to us or to init ( will be under init if unik was restarted - we try
		// make sure it's not started by someone else..)
		if proc.PPid() != os.Getpid() && proc.PPid() != 1 {
			continue
		}

		if strings.Contains(proc.Args(), instanceArg) {
			return proc, nil
		}
	}
	return nil, errors.New("Qemu process not found", nil)
}
