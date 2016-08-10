package vsphere

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func (p *VsphereProvider) CreateVolume(params types.CreateVolumeParams) (_ *types.Volume, err error) {
	if _, volumeErr := p.GetImage(params.Name); volumeErr == nil {
		return nil, errors.New("volume already exists", nil)
	}
	c := p.getClient()

	localVmdkDir, err := ioutil.TempDir("", "localvmdkdir.")
	if err != nil {
		return nil, errors.New("creating tmp file", err)
	}
	defer os.RemoveAll(localVmdkDir)
	localVmdkFile := filepath.Join(localVmdkDir, "data.vmdk")
	logrus.WithField("raw-image", params.ImagePath).Infof("creating vmdk from raw image")
	if err := common.ConvertRawImage(types.ImageFormat_RAW, types.ImageFormat_VMDK, params.ImagePath, localVmdkFile); err != nil {
		return nil, errors.New("converting raw image to vmdk", err)
	}

	rawImageFile, err := os.Stat(localVmdkFile)
	if err != nil {
		return nil, errors.New("statting raw image file", err)
	}
	sizeMb := rawImageFile.Size() >> 20

	vsphereVolumeDir := getVolumeDatastoreDir(params.Name)
	if err := c.Mkdir(vsphereVolumeDir); err != nil {
		return nil, errors.New("creating vsphere directory for volume", err)
	}
	defer func() {
		if err != nil {
			if params.NoCleanup {
				logrus.Warnf("because --no-cleanup flag was provided, not cleaning up failed volume %s at %s", params.Name, vsphereVolumeDir)
				return
			}
			logrus.WithError(err).Warnf("creating volume failed, cleaning up volume on datastore")
			c.Rmdir(vsphereVolumeDir)
		}
	}()

	if err := c.ImportVmdk(localVmdkFile, vsphereVolumeDir); err != nil {
		return nil, errors.New("importing data.vmdk to vsphere datastore", err)
	}

	volume := &types.Volume{
		Id:             params.Name,
		Name:           params.Name,
		SizeMb:         sizeMb,
		Attachment:     "",
		Infrastructure: types.Infrastructure_VSPHERE,
		Created:        time.Now(),
	}

	err = p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		volumes[volume.Id] = volume
		return nil
	})
	if err != nil {
		return nil, errors.New("modifying volume map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return nil, errors.New("saving volume map to state", err)
	}
	return volume, nil
}
