package openstack

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"time"
)

const DEFAULT_INSTANCE_DISKMB int = 10 * 1024 // 10 GB

func (p *OpenstackProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {
	// return nil, errors.New("not yet supportded for openstack", nil)

	logrus.WithFields(logrus.Fields{
		"image-id": params.ImageId,
		"mounts":   params.MntPointsToVolumeIds,
		"env":      params.Env,
	}).Infof("running instance %s", params.Name)

	clientNova, err := p.newClientNova()
	if err != nil {
		return nil, err
	}

	image, err := p.GetImage(params.ImageId)
	if err != nil {
		return nil, errors.New("failed to get image", err)
	}

	// If not set, use default.
	if params.InstanceMemory <= 0 {
		params.InstanceMemory = image.RunSpec.DefaultInstanceMemory
	}

	// Pick flavor.
	minDiskMB := image.RunSpec.MinInstanceDiskMB
	if minDiskMB <= 0 {
		// TODO(miha-plesko): raise error here, since compiler should set MinInstanceDiskMB.
		// This commit adds field MinInstanceDiskMB to the RunSpec, but ATM non of the existing
		// compilers actually set it (so it's always zero). This field should be set at compile time
		// since only then compiler is actually aware of the logical size of the disk.
		// Raise error here after compiler is updated.
		minDiskMB = DEFAULT_INSTANCE_DISKMB
	}
	flavor, err := pickFlavor(clientNova, minDiskMB, params.InstanceMemory)
	if err != nil {
		return nil, errors.New("failed to pick flavor", err)
	}

	// Run instance.
	serverId, err := launchServer(clientNova, params.Name, flavor.Name, image.Name)
	if err != nil {
		return nil, errors.New("failed to run instance", err)
	}

	instance := &types.Instance{
		Id:             serverId,
		Name:           params.Name,
		State:          types.InstanceState_Pending,
		Infrastructure: types.Infrastructure_OPENSTACK,
		ImageId:        image.Id,
		Created:        time.Now(),
	}

	// Update state.
	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		instances[instance.Id] = instance
		return nil
	}); err != nil {
		return nil, errors.New("failed to modify instance map in state", err)
	}

	logrus.WithFields(logrus.Fields{"instance": instance}).Infof("instance created succesfully")

	return instance, nil
}

// launchServer launches single server of given image and returns it's id.
func launchServer(clientNova *gophercloud.ServiceClient, name string, flavorName string, imageName string) (string, error) {
	resp := servers.Create(clientNova, servers.CreateOpts{
		Name:       name,
		FlavorName: flavorName,
		ImageName:  imageName,
	})

	if resp.Err != nil {
		return "", errors.New("failed to get OK HTTP response when running instance", resp.Err)
	}

	server, err := resp.Extract()
	return server.ID, err
}
