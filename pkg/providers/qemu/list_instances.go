package qemu

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/types"
)

func (p *QemuProvider) ListInstances() ([]*types.Instance, error) {
	if len(p.state.GetInstances()) < 1 {
		return []*types.Instance{}, nil
	}

	var instances []*types.Instance
	for _, instance := range p.state.GetInstances() {
		pid, err := strconv.Atoi(instance.Id)
		if err != nil {
			logrus.WithField("instance", instance).Warn("invalid pid - removing instance")
			p.state.RemoveInstance(instance)
			continue
		}
		if err := detectInstance(pid); err != nil {
			logrus.WithField("instance", instance).Debug("Instance is not running; removing")
			p.state.RemoveInstance(instance)
			continue
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
