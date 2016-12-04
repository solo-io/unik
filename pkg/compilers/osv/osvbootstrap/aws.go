package osvbootstrap

import (
	"io/ioutil"
	"os"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

const OSV_AWS_MEMORY = 1024

type AwsBootstrapper struct{}

func (b *AwsBootstrapper) Bootstrap(params BootstrapParams) (*types.RawImage, error) {
	// Convert to WMDK format.
	resultFile, err := ioutil.TempFile("", "osv-boot.vmdk.")
	if err != nil {
		return nil, errors.New("failed to create tmpfile for result", err)
	}
	defer func() {
		if err != nil && !params.CompileParams.NoCleanup {
			os.Remove(resultFile.Name())
		}
	}()
	if err := os.Rename(params.CapstanImagePath, resultFile.Name()); err != nil {
		return nil, errors.New("failed to rename result file", err)
	}

	return &types.RawImage{
		LocalImagePath: resultFile.Name(),
		StageSpec: types.StageSpec{
			ImageFormat:           types.ImageFormat_QCOW2,
			XenVirtualizationType: types.XenVirtualizationType_HVM,
		},
		RunSpec: types.RunSpec{
			DeviceMappings: []types.DeviceMapping{
				types.DeviceMapping{MountPoint: "/", DeviceName: "/dev/sda1"},
			},
			DefaultInstanceMemory: OSV_AWS_MEMORY,
		},
	}, nil
}

func (b *AwsBootstrapper) UseEc2() bool {
	return true
}
