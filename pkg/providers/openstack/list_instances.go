package openstack

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/pagination"
	"strings"
)

func (p *OpenstackProvider) ListInstances() ([]*types.Instance, error) {
	// Return immediately if no instance is managed by unik.
	managedInstances := p.state.GetInstances()
	if len(managedInstances) < 1 {
		return []*types.Instance{}, nil
	}

	clientNova, err := p.newClientNova()
	if err != nil {
		return nil, err
	}

	instList, err := fetchInstances(clientNova, managedInstances)
	if err != nil {
		return nil, errors.New("failed to fetch instances", err)
	}

	// Update state.
	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		// Clear everything.
		for k := range instances {
			delete(instances, k)
		}

		// Add fetched instances.
		for _, inst := range instList {
			instances[inst.Id] = inst
		}
		return nil
	}); err != nil {
		return nil, errors.New("failed to modify instance map in state", err)
	}

	return instList, nil
}

// fetchInstances fetches a list of instances runnign on OpenStack and returns a list of
// those that are managed by unik.
func fetchInstances(clientNova *gophercloud.ServiceClient, managedInstances map[string]*types.Instance) ([]*types.Instance, error) {
	var result []*types.Instance = make([]*types.Instance, 0)

	pagerServers := servers.List(clientNova, servers.ListOpts{})
	pagerServers.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}

		for _, s := range serverList {
			// Filter out instances that unik is not aware of.
			instance, ok := managedInstances[s.ID]
			if !ok {
				continue
			}

			// Interpret instance state and filter out instance with bad state.
			if state := parseInstanceState(s.Status); state == types.InstanceState_Terminated {
				continue
			} else {
				instance.State = state
			}

			// Update fields.
			instance.Name = s.Name
			instance.IpAddress = s.AccessIPv4

			result = append(result, instance)
		}

		return true, nil
	})
	return result, nil
}

func parseInstanceState(serverState string) types.InstanceState {
	// http://docs.openstack.org/developer/nova/vmstates.html#vm-states-and-possible-commands
	switch strings.ToLower(serverState) {
	case "active", "rescued":
		return types.InstanceState_Running
	case "building":
		return types.InstanceState_Pending
	case "paused":
		return types.InstanceState_Paused
	case "suspended":
		return types.InstanceState_Suspended
	case "shutoff", "stopped", "soft_deleted":
		return types.InstanceState_Stopped
	case "hard_deleted":
		return types.InstanceState_Terminated
	case "error":
		return types.InstanceState_Error
	}

	logrus.WithFields(logrus.Fields{
		"serverState": serverState,
	}).Infof("recieved unknown instance state")

	return types.InstanceState_Unknown
}
