package gcloud

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/types"
)

func (p *GcloudProvider) CreateVolume(params types.CreateVolumeParams) (*types.Volume, error) {
	return nil, errors.New("not yet implemented", nil)
}
func (p *GcloudProvider) CreateEmptyVolume(name string, size int) (*types.Volume, error) {
	return nil, errors.New("not yet implemented", nil)
}
