package aws

import (
	"mime/multipart"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxlog"
)

func (p *AwsProvider) CreateVolume(logger lxlog.Logger, name string, dataTar multipart.File, tarHeader *multipart.FileHeader, size int) (*types.Volume, error) {
	return nil, nil
}
func (p *AwsProvider) CreateEmptyVolume(logger lxlog.Logger, name string, size int) (*types.Volume, error) {
	return nil, nil
}
func (p *AwsProvider) RunInstance(logger lxlog.Logger, name, imageId string, mntPointsToVolumeIds map[string]string, env map[string]string) (*types.Instance, error) {
	return nil, nil
}
