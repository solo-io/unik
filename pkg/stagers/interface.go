package stagers

import (
	"mime/multipart"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

type Stager interface {
	//Images
	Stage(name, bootImageFile string, force bool) (*types.Image, error)
	ListImages() ([]*types.Image, error)
	GetImage(nameOrIdPrefix string) (*types.Image, error)
	DeleteImage(id string, force bool) error
	//Instances
	RunInstance(name, imageId string, mntPointsToVolumeIds map[string]string) (*types.Instance, error)
	ListInstances() ([]*types.Instance, error)
	StartInstance(id string) error
	StopInstance(id string) error
	GetInstance(nameOrIdPrefix string) (*types.Instance, error)
	DeleteInstance(id string) error
	GetLogs(id string) (string, error)
	//Volumes
	CreateVolume(name string, dataTar multipart.File, tarHeader *multipart.FileHeader) (*types.Volume, error)
	ListVolumes() ([]*types.Volume, error)
	GetVolume(nameOrIdPrefix string) (*types.Volume, error)
	DeleteVolume(id string, force bool) error
	AttachVolume(id, instanceId string) error
	DetachVolume(id string) error
}