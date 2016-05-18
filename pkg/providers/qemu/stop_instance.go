package qemu

import (
	"fmt"
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

	proc, err := getOurQemu(instance)
	if err != nil {
		return err
	}

	// kill qemu
	return syscall.Kill(proc.Pid(), syscall.SIGKILL)
}

func getOurQemu(instance *types.Instance) (ps.Process, error) {

	procs, err := ps.Processes()
	if err != nil {
		return nil, err
	}

	for _, proc := range procs {
		if !strings.Contains(proc.Executable(), "qemu") {
			continue
		}
		instanceArg := fmt.Sprintf("/instances/%s/kernel", instance.Name)
		if strings.Contains(proc.Args(), instanceArg) {
			return proc, nil
		}
	}
	return nil, errors.New("Qemu process not found", nil)
}
