// +build cgo

package firecracker

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/solo-io/unik/pkg/types"
)

func (p *FirecrackerProvider) StopInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}

	p.mapLock.RLock()
	m := p.runningMachines[id]
	p.mapLock.RUnlock()

	if m == nil {
		logrus.WithField("instance", instance).Warn("instance not available in runtime")
	} else {
		p.mapLock.Lock()
		delete(p.runningMachines, id)
		p.mapLock.Unlock()

		m.StopVMM()
	}

	volumesToDetach := []*types.Volume{}
	volumes, err := p.ListVolumes()
	if err != nil {
		return errors.New("getting volume list", err)
	}
	for _, volume := range volumes {
		if volume.Attachment == instance.Id {
			volumesToDetach = append(volumesToDetach, volume)
		}
	}

	return p.state.RemoveInstance(instance)
}
