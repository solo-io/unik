package vsphere

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/config"
	"github.com/cf-unik/unik/pkg/providers/common"
	"github.com/cf-unik/unik/pkg/providers/vsphere/vsphereclient"
	"github.com/cf-unik/unik/pkg/state"
	"time"
)

func VsphereStateFile() string {
	return filepath.Join(config.Internal.UnikHome, "vsphere/state.json")
}

var VsphereImagesDirectory = "unik/vsphere/images/"
var VsphereVolumesDirectory = "unik/vsphere/volumes/"

const VsphereUnikInstanceListener = "VsphereUnikInstanceListener"
const instanceListenerPrefix = "unik_vsphere"

type VsphereProvider struct {
	config             config.Vsphere
	state              state.State
	u                  *url.URL
	instanceListenerIp string
}

func NewVsphereProvier(config config.Vsphere) (*VsphereProvider, error) {
	rawUrl := "https://" + config.VsphereUser + ":" + config.VspherePassword + "@" + strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(config.VsphereURL, "http://"), "https://"), "/sdk") + "/sdk"
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, errors.New("parsing vsphere url", err)
	}

	p := &VsphereProvider{
		config: config,
		state:  state.NewBasicState(VsphereStateFile()),
		u:      u,
	}

	p.getClient().Mkdir("unik")
	p.getClient().Mkdir("unik/vsphere")
	p.getClient().Mkdir("unik/vsphere/images")
	p.getClient().Mkdir("unik/vsphere/volumes")

	if err := p.deployInstanceListener(); err != nil {
		return nil, errors.New("deploying virtualbox instance listener", err)
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
				logrus.Error("error updating vsphere state:", err)
			}
			time.Sleep(time.Second)
		}
	}()

	return p, nil
}

func (p *VsphereProvider) WithState(state state.State) *VsphereProvider {
	p.state = state
	return p
}

func (p *VsphereProvider) getClient() *vsphereclient.VsphereClient {
	return vsphereclient.NewVsphereClient(p.u, p.config.Datastore, p.config.Datacenter)
}

//just for consistency
func getInstanceDatastoreDir(instanceName string) string {
	return instanceName
}

func getImageDatastoreDir(imageName string) string {
	return filepath.Join(VsphereImagesDirectory, imageName+"/")
}

func getImageDatastorePath(imageName string) string {
	return filepath.Join(getImageDatastoreDir(imageName), "boot.vmdk")
}

func getVolumeDatastoreDir(volumeName string) string {
	return filepath.Join(VsphereVolumesDirectory, volumeName+"/")
}

func getVolumeDatastorePath(volumeName string) string {
	return filepath.Join(getVolumeDatastoreDir(volumeName), "data.vmdk")
}
