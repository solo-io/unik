package aws

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"io"
	"os"
	"io/ioutil"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
)

func (p *AwsProvider) CreateVolume(name string, sourceTar io.ReadCloser, size int) (*types.Volume, error) {
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
func (p *AwsProvider) CreateEmptyVolume(name string, size int) (*types.Volume, error) {
	return nil, nil
}