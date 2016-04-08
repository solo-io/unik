package aws

import (
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *AwsProvider) DeleteInstance(logger lxlog.Logger, id string) error {
	instance, err := p.GetInstance(logger, id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	param := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.Id),
		},
	}
	_, err = p.newEC2(logger).TerminateInstances(param)
	if err != nil {
		return lxerrors.New("failed to terminate instance "+instance.Id, err)
	}
	return p.State.ModifyInstances(func(instances map[string]*types.Instance) error {
		delete(instances, instance.Id)
		return nil
	})
}