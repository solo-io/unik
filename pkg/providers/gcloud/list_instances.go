package gcloud

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/types"
)

func (p *GcloudProvider) ListInstances() ([]*types.Instance, error) {
	if len(p.state.GetInstances()) < 1 {
		return []*types.Instance{}, nil
	}

	gInstances, err := p.compute().Instances.List(p.config.ProjectID, p.config.Zone).Do()
	if err != nil {
		return nil, errors.New("getting instance list from gcloud", err)
	}

	updatedInstances := []*types.Instance{}
	for _, instance := range p.state.GetInstances() {
		instanceFound := false
		//find instance in list
		for _, gInstance := range gInstances.Items {
			if gInstance.Name == instance.Name {
				instance.State = parseInstanceState(gInstance.Status)

				//use first network interface, skip if unavailable
				if len(gInstance.NetworkInterfaces) > 0 && len(gInstance.NetworkInterfaces[0].AccessConfigs) > 0 {
					instance.IpAddress = gInstance.NetworkInterfaces[0].AccessConfigs[0].NatIP
				}
				p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
					instances[instance.Id] = instance
					return nil
				})
				updatedInstances = append(updatedInstances, instance)
				instanceFound = true
				break
			}
		}
		if !instanceFound {
			logrus.Warnf("instance %v no longer found, cleaning it from state", instance.Name)
			p.state.RemoveInstance(instance)
		}
	}

	return updatedInstances, nil
}

func parseInstanceState(status string) types.InstanceState {
	switch status {
	case "RUNNING":
		return types.InstanceState_Running
	case "PROVISIONING":
		fallthrough
	case "STAGING":
		return types.InstanceState_Pending
	case "SUSPENDED":
		fallthrough
	case "STOPPING":
		fallthrough
	case "SUSPENDING":
		fallthrough
	case "STOPPED":
		return types.InstanceState_Stopped
	case "TERMINATED":
		return types.InstanceState_Terminated
	}
	return types.InstanceState_Unknown
}
