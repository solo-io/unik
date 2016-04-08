package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxlog"
)

func (p *AwsProvider) DeleteImage(logger lxlog.Logger, id string, force bool) error {
	image, err := p.GetImage(logger, id)
	if err != nil {
		return lxerrors.New("retrieving image", err)
	}
	instances, err := p.ListInstances(logger)
	if err != nil {
		return lxerrors.New("retrieving list of instances", err)
	}
	for _, instance := range instances {
		if instance.ImageId == image.Id {
			if !force {
				return lxerrors.New("instance "+instance.Id+" found which uses image "+image.Id+"; try again with force=true", nil)
			} else {
				err = p.DeleteInstance(logger, instance.Id)
				if err != nil{
					return lxerrors.New("failed to delete instance "+instance.Id+" which is using image "+image.Id, err)
				}
			}
		}
	}

	svc := p.newEC2(logger)
	svc.ImportImage()
	describeSnapshotsOutput, err := svc.DescribeSnapshots(&ec2.DescribeSnapshotsInput{})
	if err != nil {
		return lxerrors.New("getting ec2 snapshot list", err)
	}
	snap, err := getSnapshotForImage(describeSnapshotsOutput)
	if err != nil {
		return err
	}

	snap.VolumeId
}

make sure we tag the snapshot when we tag the ami
func getSnapshotForImage(describeSnapshotsOutput *ec2.DescribeSnapshotsOutput, imageId string) (*ec2.Snapshot, error) {
	for _, snapshot := range describeSnapshotsOutput.Snapshots {
		for _, tag := range snapshot.Tags {
			if *tag.Key == UNIK_IMAGE_ID && *tag.Value == imageId {
				return snapshot, nil
			}
		}
	}
	return nil, lxerrors.New("snapshot for image "+imageId+" not found", nil)
}