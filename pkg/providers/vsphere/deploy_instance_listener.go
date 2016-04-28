package vsphere

import (
	"github.com/Sirupsen/logrus"
	"github.com/layer-x/layerx-commons/lxerrors"
	"io"
	"net/http"
	"os"
	"strings"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
)

const (
	vsphereInstanceListenerUrl = "https://s3.amazonaws.com/unik-instance-listener/vsphere-instancelistener-base.vmdk"
	vsphereInstanceListenerVmdk = "vsphere-instancelistener-base.vmdk"
)

func (p *VsphereProvider) DeployInstanceListener() error {
	logrus.Infof("checking if instance listener base vmdk already exists on vsphere datastore")
	c := p.getClient()
	vm, err := c.GetVm(VsphereUnikInstanceListener)
	if err == nil {
		if vm.Summary.Runtime.PowerState != "poweredOn" {
			return c.PowerOnVm(VsphereUnikInstanceListener)
		} else {
			logrus.Info("instance listener already running")
			return nil
		}
	}
	files, err := c.Ls("unik")
	if err != nil {
		return lxerrors.New("lsing on folder 'unik'", err)
	}
	alreadyUploaded := false
	for _, file := range files {
		if strings.Contains(file, vsphereInstanceListenerVmdk) {
			alreadyUploaded = true
			break
		}
	}
	if !alreadyUploaded {
		if _, err := os.Stat(vsphereInstanceListenerVmdk); err != nil {
			logrus.WithError(err).Infof("vsphere instance listener vmdk not found, attempting to download from " + vsphereInstanceListenerUrl)
			vmdkFile, err := os.Create(vsphereInstanceListenerVmdk)
			if err != nil {
				return lxerrors.New("creating file for vsphere instance listener vmdk", err)
			}
			resp, err := http.Get(vsphereInstanceListenerUrl)
			if err != nil {
				return lxerrors.New("contacting "+ vsphereInstanceListenerUrl, err)
			}
			defer resp.Body.Close()
			n, err := io.Copy(vmdkFile, unikutil.ReaderWithProgress(resp.Body, resp.ContentLength))
			if err != nil {
				return lxerrors.New("copying response to file", err)
			}
			logrus.Infof("%v bytes written", n)
		}
		logrus.Infof("uploading " + vsphereInstanceListenerVmdk)
		if err := c.ImportVmdk(vsphereInstanceListenerVmdk, "unik/"+vsphereInstanceListenerVmdk); err != nil {
			return lxerrors.New("copying instance listener vmdk", err)
		}
	}

	logrus.Infof("deploying vsphere instance listener")
	if err := c.CreateVm(VsphereUnikInstanceListener, 512); err != nil {
		return lxerrors.New("creating vm", err)
	}
	if err := c.AttachDisk(VsphereUnikInstanceListener, "unik/"+vsphereInstanceListenerVmdk, 0); err != nil {
		return lxerrors.New("attaching disk to vm", err)
	}
	if err := c.PowerOnVm(VsphereUnikInstanceListener); err != nil {
		return lxerrors.New("powering on vm", err)
	}
	return nil
}
