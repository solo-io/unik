package aws

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *AwsProvider) ListInstances() ([]*types.Instance, error) {
	if len(p.state.GetInstances()) < 1 {
		return []*types.Instance{}, nil
	}

	instanceIds := []*string{}
	for instanceId := range p.state.GetInstances() {
		instanceIds = append(instanceIds, aws.String(instanceId))
	}
	param := &ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	}
	output, err := p.newEC2().DescribeInstances(param)
	if err != nil {
		return nil, errors.New("running ec2 describe instances ", err)
	}
	updatedInstances := []*types.Instance{}
	for _, reservation := range output.Reservations {
		for _, ec2Instance := range reservation.Instances {
			logrus.WithField("ec2instance", ec2Instance).Debugf("aws returned instance %s", *ec2Instance.InstanceId)
			var instanceId string
			if ec2Instance.InstanceId != nil {
				instanceId = *ec2Instance.InstanceId
			}
			if instanceId == "" {
				logrus.Warnf("instance %v does not have readable instanceId, moving on", *ec2Instance)
				continue
			}
			instanceState := parseInstanceState(ec2Instance.State)
			if instanceState == types.InstanceState_Unknown {
				logrus.Warnf("instance %s state is unknown (%s), moving on", instanceId, *ec2Instance.State.Name)
				continue
			}
			if instanceState == types.InstanceState_Terminated {
				logrus.Warnf("instance %s state is terminated, removing it from state", instanceId)
				err = p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
					delete(instances, instanceId)
					return nil
				})
				if err != nil {
					return nil, errors.New("modifying instance map in state", err)
				}
				err = p.state.Save()
				if err != nil {
					return nil, errors.New("saving modified instance map to state", err)
				}
				continue
			}
			instance, ok := p.state.GetInstances()[instanceId]
			if !ok {
				logrus.WithFields(logrus.Fields{"ec2Instance": ec2Instance}).Errorf("found an instance that unik has no record of")
				continue
			}
			instance.State = instanceState
			if ec2Instance.PublicIpAddress != nil {
				instance.IpAddress = *ec2Instance.PublicIpAddress
			}
			err = p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
				instances[instance.Id] = instance
				return nil
			})
			if err != nil {
				return nil, errors.New("modifying instance map in state", err)
			}
			err = p.state.Save()
			if err != nil {
				return nil, errors.New("saving modified instance map to state", err)
			}
			updatedInstances = append(updatedInstances, instance)
		}
	}
	return updatedInstances, nil
}

func parseInstanceState(ec2State *ec2.InstanceState) types.InstanceState {
	if ec2State == nil {
		return types.InstanceState_Unknown
	}
	switch *ec2State.Name {
	case ec2.InstanceStateNameRunning:
		return types.InstanceState_Running
	case ec2.InstanceStateNamePending:
		return types.InstanceState_Pending
	case ec2.InstanceStateNameStopped:
		return types.InstanceState_Stopped
	case ec2.InstanceStateNameTerminated:
		return types.InstanceState_Terminated
	}
	return types.InstanceState_Unknown
}
