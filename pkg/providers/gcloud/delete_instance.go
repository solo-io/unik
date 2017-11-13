package gcloud

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/types"
)

func (p *GcloudProvider) DeleteInstance(id string, force bool) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	if instance.State == types.InstanceState_Running && !force {
		return errors.New("instance "+instance.Id+"is still running. try again with --force or power off instance first", err)
	}
	_, err = p.compute().Instances.Delete(p.config.ProjectID, p.config.Zone, instance.Name).Do()
	if err != nil {
		return errors.New("failed to terminate instance "+instance.Id, err)
	}
	return p.state.RemoveInstance(instance)
}
