package firecracker

import (
	"os"

	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *FirecrackerProvider) DeleteVolume(id string, force bool) error {

	volume, err := p.GetVolume(id)
	if err != nil {
		return errors.New("retrieving volume "+id, err)
	}
	if volume.Attachment != "" {
		if force {
			if err := p.DetachVolume(volume.Id); err != nil {
				return errors.New("detaching volume for deletion", err)
			}
		} else {
			return errors.New("volume "+volume.Id+" is attached to instance."+volume.Attachment+", try again with --force or detach volume first", err)
		}
	}
	volumePath := getVolumePath(volume.Name)
	err = os.Remove(volumePath)
	if err != nil {
		return errors.New("could not delete volume at path "+volumePath, err)
	}
	return p.state.RemoveVolume(volume)
}
