package aws

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxlog"
	"io"
	"os"
	"io/ioutil"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
)

func (p *AwsProvider) CreateVolume(logger lxlog.Logger, name string, sourceTar io.ReadCloser, size int) (*types.Volume, error) {
	localFolder, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(localFolder)

	if err := unikos.ExtractTar(sourceTar, localFolder); err != nil {
		return nil, err
	}
	return nil, nil
}
func (p *AwsProvider) CreateEmptyVolume(logger lxlog.Logger, name string, size int) (*types.Volume, error) {
	return nil, nil
}