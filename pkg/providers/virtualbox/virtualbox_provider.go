package virtualbox

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"time"
)

func VirtualboxStateFile() string {
	return filepath.Join(config.Internal.UnikHome, "virtualbox/state.json")
}
func virtualboxImagesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "virtualbox/images/")
}
func virtualboxInstancesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "virtualbox/instances/")
}
func virtualboxVolumesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "virtualbox/volumes/")
}

const VboxUnikInstanceListener = "VboxUnikInstanceListener"
const instanceListenerPrefix = "unik_virtualbox"

type VirtualboxProvider struct {
	config             config.Virtualbox
	state              state.State
	instanceListenerIp string
}

func NewVirtualboxProvider(config config.Virtualbox) (*VirtualboxProvider, error) {
	os.MkdirAll(virtualboxImagesDirectory(), 0755)
	os.MkdirAll(virtualboxInstancesDirectory(), 0755)
	os.MkdirAll(virtualboxVolumesDirectory(), 0755)

	p := &VirtualboxProvider{
		config: config,
		state:  state.NewBasicState(VirtualboxStateFile()),
	}

	if err := p.deployInstanceListener(config); err != nil && !strings.Contains(err.Error(), "already exists") {
		return nil, errors.New("deploing virtualbox instance listener", err)
	}

	instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, timeout)
	if err != nil {
		return nil, errors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
	}

	p.instanceListenerIp = instanceListenerIp

	// begin update instances cycle
	go func() {
		for {
			if err := p.syncState(); err != nil {
				logrus.Error("error updatin virtualbox state:", err)
			}
			time.Sleep(time.Second)
		}
	}()

	return p, nil
}

func (p *VirtualboxProvider) WithState(state state.State) *VirtualboxProvider {
	p.state = state
	return p
}

func getImagePath(imageName string) string {
	return filepath.Join(virtualboxImagesDirectory(), imageName, "boot.vmdk")
}

func getInstanceDir(instanceName string) string {
	return filepath.Join(virtualboxInstancesDirectory(), instanceName)
}

func getVolumePath(volumeName string) string {
	return filepath.Join(virtualboxVolumesDirectory(), volumeName, "data.vmdk")
}
