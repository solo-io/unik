package ukvm

import (
	"fmt"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"os"
	"strconv"
	"syscall"
)

func (p *UkvmProvider) ListInstances() ([]*types.Instance, error) {
	if len(p.state.GetInstances()) < 1 {
		return []*types.Instance{}, nil
	}

	var instances []*types.Instance
	for _, instance := range p.state.GetInstances() {
		pid, err := strconv.Atoi(instance.Id)
		if err != nil {
			return nil, errors.New("invalid id (is not a pid)", err)
		}
		if err := detectInstance(pid); err != nil {
			p.state.RemoveInstance(instance)
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

func detectInstance(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return errors.New("Failed to find process", err)
	}
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return errors.New(fmt.Sprintf("process.Signal on pid %d returned", pid), err)
	}
	return nil
}
