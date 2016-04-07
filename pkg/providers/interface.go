package providers

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxlog"
	"mime/multipart"
)

type Provider interface {
	//Images
	Stage(logger lxlog.Logger, name string, rawImage *types.RawImage, force bool) (*types.Image, error)
	ListImages(logger lxlog.Logger) ([]*types.Image, error)
	GetImage(logger lxlog.Logger, nameOrIdPrefix string) (*types.Image, error)
	DeleteImage(logger lxlog.Logger, id string, force bool) error
	//Instances
	RunInstance(logger lxlog.Logger, name, imageId string, mntPointsToVolumeIds map[string]string, env map[string]string) (*types.Instance, error)
	ListInstances(logger lxlog.Logger) ([]*types.Instance, error)
	StartInstance(logger lxlog.Logger, id string) error
	StopInstance(logger lxlog.Logger, id string) error
	GetInstance(logger lxlog.Logger, nameOrIdPrefix string) (*types.Instance, error)
	DeleteInstance(logger lxlog.Logger, id string) error
	GetLogs(logger lxlog.Logger, id string) (string, error)
	//Volumes
	CreateVolume(logger lxlog.Logger, name string, dataTar multipart.File, tarHeader *multipart.FileHeader, size int) (*types.Volume, error)
	CreateEmptyVolume(logger lxlog.Logger, name string, size int) (*types.Volume, error)
	ListVolumes(logger lxlog.Logger) ([]*types.Volume, error)
	GetVolume(logger lxlog.Logger, nameOrIdPrefix string) (*types.Volume, error)
	DeleteVolume(logger lxlog.Logger, id string, force bool) error
	AttachVolume(logger lxlog.Logger, id, instanceId string) error
	DetachVolume(logger lxlog.Logger, id string) error
}
