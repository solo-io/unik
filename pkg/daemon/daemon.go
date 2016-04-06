package daemon

import (
	"github.com/go-martini/martini"
	"github.com/layer-x/layerx-commons/lxmartini"
	"github.com/layer-x/unik/pkg/daemon/vsphere"
	"github.com/layer-x/layerx-commons/lxlog"
)

type UnikDaemon struct {
	server *martini.ClassicMartini
	cpi    UnikCPI
}

func NewUnikDaemon(provider string, opts map[string]string) *UnikDaemon {
	logger := lxlog.New("daemon-setup")
	var cpi UnikCPI
	switch provider{
	case "ec2":
		cpi = ec2.NewUnikEC2CPI()
		break
	case "vsphere":
		vsphereCpi := vsphere.NewUnikVsphereCPI(logger, opts["vsphereUrl"], opts["vsphereUser"], opts["vspherePass"])
		vsphereCpi.StartInstanceDiscovery(logger)
		vsphereCpi.ListenForBootstrap(logger, 3001)
		cpi = vsphereCpi
		break
	default:
		logger.Fatalf("Unrecognized provider " + provider)
	}
	return &UnikDaemon{
		server: lxmartini.QuietMartini(),
		cpi: cpi,
	}
}