package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/types"
)

func (p *AwsProvider) DeleteInstance(id string, force bool) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	if instance.State == types.InstanceState_Running && !force {
		return errors.New("instance "+instance.Id+"is still running. try again with --force or power off instance first", err)
	}
	param := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.Id),
		},
	}
	_, err = p.newEC2().TerminateInstances(param)
	if err != nil {
		return errors.New("failed to terminate instance "+instance.Id, err)
	}
	return p.state.RemoveInstance(instance)
}
