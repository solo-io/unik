package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxlog"
)

func (p *AwsProvider) StopInstance(logger lxlog.Logger, id string) error {
	instance, err := p.GetInstance(logger, id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	param := &ec2.StopInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.Id),
		},
	}
	_, err = p.newEC2(logger).StopInstances(param)
	if err != nil {
		return lxerrors.New("failed to stop instance "+instance.Id, err)
	}
	return nil
}
