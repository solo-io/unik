package gcloud

import (
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *GcloudProvider) StartInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	if _, err := p.compute().Instances.Start(p.config.ProjectID, p.config.Zone, instance.Name).Do(); err != nil {
		return errors.New("failed to start instance "+instance.Id, err)
	}
	return nil
}
