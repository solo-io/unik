package virtualbox

import (
	"github.com/Sirupsen/logrus"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/layer-x/layerx-commons/lxerrors"
	"io"
	"net/http"
	"os"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
)

const (
	vboxInstanceListenerUrl  = "https://s3.amazonaws.com/unik-instance-listener/vbox-instancelistener-base.vmdk"
	vboxInstanceListenerVmdk = "instancelistener-base.vmdk"
)

func (p *VirtualboxProvider) DeployInstanceListener() error {
	if _, err := os.Stat(vboxInstanceListenerVmdk); err != nil {
		logrus.Infof("vbox instance listener vmdk not found, attempting to download from " + vboxInstanceListenerUrl)
		vmdkFile, err := os.Create(vboxInstanceListenerVmdk)
		if err != nil {
			return lxerrors.New("creating file for vbox instance listener vmdk", err)
		}
		resp, err := http.Get(vboxInstanceListenerUrl)
		if err != nil {
			return lxerrors.New("contacting "+vboxInstanceListenerUrl, err)
		}
		defer resp.Body.Close()
		n, err := io.Copy(vmdkFile, unikutil.ReaderWithProgress(resp.Body, resp.ContentLength))
		if err != nil {
			return lxerrors.New("copying response to file", err)
		}
		logrus.Infof("%v bytes written"+vboxInstanceListenerUrl, n)
	}

	logrus.Infof("deploying virtualbox instance listener")
	if err := virtualboxclient.CreateVmNatless(VboxUnikInstanceListener, os.Getenv("PWD"), p.config.AdapterName, p.config.VirtualboxAdapterType); err != nil {
		return lxerrors.New("creating vm", err)
	}
	if err := unikos.CopyFile(vboxInstanceListenerVmdk, "instancelistener-copy.vmdk"); err != nil {
		return lxerrors.New("copying instance listener vmdk", err)
	}
	if err := virtualboxclient.AttachDisk(VboxUnikInstanceListener, "instancelistener-copy.vmdk", 0); err != nil {
		return lxerrors.New("attaching disk to vm", err)
	}
	if err := virtualboxclient.PowerOnVm(VboxUnikInstanceListener); err != nil {
		return lxerrors.New("powering on vm", err)
	}
	return nil
}
