package gcloud

import (
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *GcloudProvider) StopInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	if _, err := p.compute().Instances.Stop(p.config.ProjectID, p.config.Zone, instance.Name).Do(); err != nil {
		return errors.New("failed to stop instance "+instance.Id, err)
	}
	return nil
}
