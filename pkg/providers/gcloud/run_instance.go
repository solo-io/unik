package gcloud

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"google.golang.org/api/compute/v1"
	"time"
)

func (p *GcloudProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {
	logrus.WithFields(logrus.Fields{
		"image-id": params.ImageId,
		"mounts":   params.MntPointsToVolumeIds,
		"env":      params.Env,
	}).Infof("running instance %s", params.Name)

	var instanceId string

	defer func() {
		if err != nil {
			logrus.WithError(err).Errorf("gcloud running instance encountered an error")
			if instanceId != "" {
				if params.NoCleanup {
					logrus.Warnf("because --no-cleanup flag was provided, not cleaning up failed instance %s0", instanceId)
					return
				}
				logrus.Warnf("cleaning up instance %s", instanceId)
				p.compute().Instances.Delete(p.config.ProjectID, p.config.Zone, instanceId)
				if cleanupErr := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
					delete(instances, instanceId)
					return nil
				}); cleanupErr != nil {
					logrus.Error(errors.New("modifying instance map in state", cleanupErr))
				}
			}
		}
	}()

	image, err := p.GetImage(params.ImageId)
	if err != nil {
		return nil, errors.New("getting image", err)
	}

	if err := common.VerifyMntsInput(p, image, params.MntPointsToVolumeIds); err != nil {
		return nil, errors.New("invalid mapping for volume", err)
	}

	envData, err := json.Marshal(params.Env)
	if err != nil {
		return nil, errors.New("could not convert instance env to json", err)
	}

	//if not set, use default
	if params.InstanceMemory <= 0 {
		params.InstanceMemory = image.RunSpec.DefaultInstanceMemory
	}

	if len(envData) > 32768 {
		return nil, errors.New("total length of env metadata must be <= 32768 bytes; have json string "+string(envData), nil)
	}

	disks := []*compute.AttachedDisk{
		//boot disk
		&compute.AttachedDisk{
			AutoDelete: true,
			Boot:       true,
			//DeviceName: "sd0"
			InitializeParams: &compute.AttachedDiskInitializeParams{
				SourceImage: "global/images/" + image.Name,
			},
		},
	}

	for _, volumeId := range params.MntPointsToVolumeIds {
		disks = append(disks, &compute.AttachedDisk{
			AutoDelete: false,
			Boot:       false,
			Source:     volumeId,
		})
	}

	instanceSpec := &compute.Instance{
		Name: params.Name,
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				&compute.MetadataItems{
					Key:   "ENV_DATA",
					Value: pointerTo(string(envData)),
				},
			},
		},
		Disks:       disks,
		MachineType: fmt.Sprintf("zones/%s/machineTypes/%s", p.config.Zone, "g1-small"),
		NetworkInterfaces: []*compute.NetworkInterface{
			&compute.NetworkInterface{
				AccessConfigs: []*compute.AccessConfig{
					&compute.AccessConfig{
						Type: "ONE_TO_ONE_NAT",
						Name: "External NAT",
					},
				},
				Network: "global/networks/default",
			},
		},
	}

	gInstance, err := p.compute().Instances.Insert(p.config.ProjectID, p.config.Zone, instanceSpec).Do()
	if err != nil {
		return nil, errors.New("creating instance on gcloud failed", err)
	}
	logrus.Infof("gcloud instance created: %+v", gInstance)

	instanceId = params.Name

	//must add instance to state before attaching volumes
	instance := &types.Instance{
		Id:             instanceId,
		Name:           params.Name,
		State:          types.InstanceState_Pending,
		Infrastructure: types.Infrastructure_GCLOUD,
		ImageId:        image.Id,
		Created:        time.Now(),
	}

	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		instances[instance.Id] = instance
		return nil
	}); err != nil {
		return nil, errors.New("modifying instance map in state", err)
	}

	logrus.WithFields(logrus.Fields{"instance": instance}).Infof("instance created succesfully")

	return instance, nil
}

func pointerTo(v string) *string {
	return &v
}
