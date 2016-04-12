package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/Sirupsen/logrus"
)

const UNIK_INSTANCE_ID = "UNIK_INSTANCE_ID"

func (p *AwsProvider) ListInstances() ([]*types.Instance, error) {
	param := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("tag-key"),
				Values: []*string{aws.String(UNIK_INSTANCE_ID)},
			},
		},
	}
	output, err := p.newEC2().DescribeInstances(param)
	if err != nil {
		return nil, lxerrors.New("running ec2 describe instances ", err)
	}
	updatedInstances := []*types.Instance{}
	for _, reservation := range output.Reservations {
		for _, ec2Instance := range reservation.Instances {
			instanceId := parseInstanceId(ec2Instance)
			if instanceId == "" {
				continue
			}
			instance, ok := p.state.GetInstances()[instanceId]
			if !ok {
				logrus.WithFields(logrus.Fields{"ec2Instance": ec2Instance}).Errorf("found an instance that unik has no record of")
				continue
			}
			instance.State = parseInstanceState(ec2Instance.State)
			instance.IpAddress = *ec2Instance.PublicIpAddress
			p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
				instances[instance.Id] = instance
				return nil
			})
			updatedInstances = append(updatedInstances, instance)
		}
	}
	return updatedInstances, nil
}

func parseInstanceId(instance *ec2.Instance) string {
	for _, tag := range instance.Tags {
		if *tag.Key == UNIK_INSTANCE_ID {
			return *tag.Value
		}
	}
	return ""
}

func parseInstanceState(ec2State *ec2.InstanceState) types.InstanceState {
	if ec2State == nil {
		return types.InstanceState_Unknown
	}
	switch *ec2State.Name {
	case "running":
		return types.InstanceState_Running
	case "pending":
		return types.InstanceState_Pending
	case "stopped":
		return types.InstanceState_Stopped
	case "shutting-down":
		return types.InstanceState_Stopped
	case "terminated":
		return types.InstanceState_Terminating
	}
	return types.InstanceState_Unknown
}
