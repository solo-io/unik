package virtualbox

import (
	"github.com/Sirupsen/logrus"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/pkg/errors"
	"io"
	"net/http"
	"os"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/emc-advanced-dev/unik/pkg/config"
)

const (
	vboxInstanceListenerUrl  = "https://s3.amazonaws.com/unik-instance-listener/vbox-instancelistener-base.vmdk"
	vboxInstanceListenerVmdk = "vbox-instancelistener-base.vmdk"
)

func (p *VirtualboxProvider) DeployInstanceListener(config config.Virtualbox) error {
	if _, err := virtualboxclient.GetVm(VboxUnikInstanceListener); err != nil {
		logrus.Warnf(VboxUnikInstanceListener+" not found! Beginning deploy...")
	} else {
		virtualboxclient.PowerOffVm(VboxUnikInstanceListener)
		virtualboxclient.ConfigureVmNetwork(VboxUnikInstanceListener, config.AdapterName, config.VirtualboxAdapterType)
		virtualboxclient.PowerOnVm(VboxUnikInstanceListener)
		return nil
	}

	if _, err := os.Stat(vboxInstanceListenerVmdk); err != nil {
		logrus.WithError(err).Infof("vbox instance listener vmdk not found, attempting to download from " + vboxInstanceListenerUrl)
		vmdkFile, err := os.Create(vboxInstanceListenerVmdk)
		if err != nil {
			return errors.New("creating file for vbox instance listener vmdk", err)
		}
		resp, err := http.Get(vboxInstanceListenerUrl)
		if err != nil {
			return errors.New("contacting "+vboxInstanceListenerUrl, err)
		}
		defer resp.Body.Close()
		n, err := io.Copy(vmdkFile, unikutil.ReaderWithProgress(resp.Body, resp.ContentLength))
		if err != nil {
			return errors.New("copying response to file", err)
		}
		logrus.Infof("%v bytes written"+vboxInstanceListenerUrl, n)
	}

	logrus.Infof("deploying virtualbox instance listener")
	if err := virtualboxclient.CreateVmNatless(VboxUnikInstanceListener, os.Getenv("PWD"), p.config.AdapterName, p.config.VirtualboxAdapterType); err != nil {
		return errors.New("creating vm", err)
	}
	if err := unikos.CopyFile(vboxInstanceListenerVmdk, "vbox-instancelistener-copy.vmdk"); err != nil {
		return errors.New("copying instance listener vmdk", err)
	}
	if err := virtualboxclient.AttachDisk(VboxUnikInstanceListener, "vbox-instancelistener-copy.vmdk", 0); err != nil {
		return errors.New("attaching disk to vm", err)
	}
	if err := virtualboxclient.PowerOnVm(VboxUnikInstanceListener); err != nil {
		return errors.New("powering on vm", err)
	}
	return nil
}
