package openstack

import (
	"fmt"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

func (p *OpenstackProvider) DeleteInstance(id string, force bool) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to retrieve instance '%s'", id), err)
	}
	if instance.State == types.InstanceState_Running && !force {
		return errors.New(fmt.Sprintf("instance '%s' is still running. try again with --force or power off instance first", instance.Id), err)
	}

	clientNova, err := p.newClientNova()
	if err != nil {
		return err
	}

	if deleteErr := deleteServer(clientNova, instance.Id); deleteErr != nil {
		return errors.New(fmt.Sprintf("failed to terminate instance '%s'", instance.Id), deleteErr)
	}

	// Update state.
	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		delete(instances, instance.Id)
		return nil
	}); err != nil {
		return errors.New("failed to modify image map in state", err)
	}
	return nil
}

func deleteServer(clientNova *gophercloud.ServiceClient, instanceId string) error {
	return servers.Delete(clientNova, instanceId).Err
}
