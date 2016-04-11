package providers

import (
	"errors"
	"mime/multipart"

	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxlog"
)

type Provider interface {
	//Images
	Stage(logger lxlog.Logger, name string, compileFunc func() (*types.RawImage, error), force bool) (*types.Image, error)
	ListImages(logger lxlog.Logger) ([]*types.Image, error)
	GetImage(logger lxlog.Logger, nameOrIdPrefix string) (*types.Image, error)
	DeleteImage(logger lxlog.Logger, id string, force bool) error
	//Instances
	RunInstance(logger lxlog.Logger, name, imageId string, mntPointsToVolumeIds map[string]string, env map[string]string) (*types.Instance, error)
	ListInstances(logger lxlog.Logger) ([]*types.Instance, error)
	GetInstance(logger lxlog.Logger, nameOrIdPrefix string) (*types.Instance, error)
	DeleteInstance(logger lxlog.Logger, id string) error
	StartInstance(logger lxlog.Logger, id string) error
	StopInstance(logger lxlog.Logger, id string) error
	GetInstanceLogs(logger lxlog.Logger, id string) (string, error)
	//Volumes
	CreateVolume(logger lxlog.Logger, name string, dataTar multipart.File, tarHeader *multipart.FileHeader, size int) (*types.Volume, error)
	CreateEmptyVolume(logger lxlog.Logger, name string, size int) (*types.Volume, error)
	ListVolumes(logger lxlog.Logger) ([]*types.Volume, error)
	GetVolume(logger lxlog.Logger, nameOrIdPrefix string) (*types.Volume, error)
	DeleteVolume(logger lxlog.Logger, id string, force bool) error
	AttachVolume(logger lxlog.Logger, id, instanceId, mntPoint string) error
	DetachVolume(logger lxlog.Logger, id string) error
}

type Providers map[string]Provider

func (providers Providers) Keys() []string {
	keys := []string{}
	for providerType := range providers {
		keys = append(keys, providerType)
	}
	return keys
}

func (providers Providers) ProviderForImage(logger lxlog.Logger, imageId string) (Provider, error) {
	for _, provider := range providers {
		_, err := provider.GetImage(logger, imageId)
		if err == nil {
			return provider, nil
		}
	}
	return nil, errors.New("provider not found for image " + imageId)
}

func (providers Providers) ProviderForInstance(logger lxlog.Logger, instanceId string) (Provider, error) {
	for _, provider := range providers {
		_, err := provider.GetInstance(logger, instanceId)
		if err == nil {
			return provider, nil
		}
	}
	return nil, errors.New("provider not found for instance " + instanceId)
}

func (providers Providers) ProviderForVolume(logger lxlog.Logger, volumeId string) (Provider, error) {
	for _, provider := range providers {
		_, err := provider.GetVolume(logger, volumeId)
		if err == nil {
			return provider, nil
		}
	}
	return nil, errors.New("provider not found for volume " + volumeId)
}
