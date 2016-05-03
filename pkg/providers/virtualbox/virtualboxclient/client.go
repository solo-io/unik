package virtualboxclient

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/layer-x/layerx-commons/lxerrors"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type VboxVm struct {
	Name    string
	UUID    string
	MACAddr string
	Devices []*VboxDevice
	Running bool
}

type VboxDevice struct {
	DiskFile      string
	ControllerKey string
}

func (vm *VboxVm) String() string {
	if vm == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%-v", *vm)
}

func vboxManage(args ...string) ([]byte, error) {
	cmd := exec.Command("VBoxManage", args...)
	logrus.WithField("command", cmd.Args).Debugf("running VBoxManage command")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", string(out))
	}
	//logrus.WithField("vbox-manage-command-result", string(out)).Debugf("VBoxManage result")
	return out, nil
}

//for vms with virtualbox guest additions
func GetVmIp(vmName string) (string, error) {
	out, err := vboxManage("guestproperty", "get", vmName, "/VirtualBox/GuestInfo/Net/0/V4/IP")
	if err != nil {
		return "", lxerrors.New("retrieving vm ip", err)
	}
	if strings.Contains(string(out), "No value set") {
		return "", lxerrors.New("ip property not available for this vm", nil)
	}
	r, err := regexp.Compile("([0-9]{1,3}[\\.]){3}[0-9]{1,3}")
	if err != nil {
		return "", lxerrors.New("compiling regex", err)
	}
	ipAddr := r.Find(out)
	if ipAddr == nil {
		return "", lxerrors.New("ip address not found in string "+string(out), nil)
	}
	return string(ipAddr), nil
}

func parseVmInfo(vmInfo string) (*VboxVm, error) {
	var uuid, macAddr string
	var running bool
	devices := []*VboxDevice{}
	lines := strings.Split(vmInfo, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "UUID:") {
			rLineBegin, err := regexp.Compile("UUID:\\ +")
			if err != nil {
				return nil, lxerrors.New("compiling regex", err)
			}
			uuid = string(rLineBegin.ReplaceAll([]byte(line), []byte("")))
		}
		if strings.Contains(line, "NIC 1:") { //first network adapter must be the IP we use
			rLineBegin, err := regexp.Compile("NIC 1:.*MAC. ")
			if err != nil {
				return nil, lxerrors.New("compiling regex", err)
			}
			rLineEnd, err := regexp.Compile(",.*")
			if err != nil {
				return nil, lxerrors.New("compiling regex", err)
			}
			macAddr = formatMac(string(rLineBegin.ReplaceAll(rLineEnd.ReplaceAll([]byte(line), []byte("")), []byte(""))))
			logrus.Debugf("mac address found for vm: %s", macAddr)
		}
		if strings.Contains(line, "SCSI (") {
			device, err := parseDevice(line)
			if err == nil {
				devices = append(devices, device)
			}
		}
		if strings.Contains(line, "State") && strings.Contains(line, "running") {
			running = true
		}
	}
	if macAddr == "" {
		return nil, lxerrors.New("mac address not found in vm info: "+string(vmInfo), nil)
	}
	if uuid == "" {
		return nil, lxerrors.New("uuid address not found in vm info: "+string(vmInfo), nil)
	}
	return &VboxVm{MACAddr: macAddr, Running: running, Devices: devices, UUID: uuid}, nil
}

func parseDevice(deviceLine string) (*VboxDevice, error) {
	rLineBegin, err := regexp.Compile("SCSI \\([0-9], [0-15]\\): ")
	if err != nil {
		return nil, lxerrors.New("compiling regex", err)
	}
	rLineEnd, err := regexp.Compile("\\(UUID: .*")
	if err != nil {
		return nil, lxerrors.New("compiling regex", err)
	}
	diskFile := rLineBegin.ReplaceAll(rLineEnd.ReplaceAll([]byte(deviceLine), []byte("")), []byte(""))
	rLineBegin, err = regexp.Compile(".*\\([0-15], ")
	if err != nil {
		return nil, lxerrors.New("compiling regex", err)
	}
	rLineEnd, err = regexp.Compile("\\):.*")
	if err != nil {
		return nil, lxerrors.New("compiling regex", err)
	}
	controllerKey := rLineBegin.ReplaceAll(rLineEnd.ReplaceAll([]byte(deviceLine), []byte("")), []byte(""))
	return &VboxDevice{DiskFile: string(diskFile), ControllerKey: string(controllerKey)}, nil
}

func Vms() ([]*VboxVm, error) {
	out, err := vboxManage("list", "vms")
	if err != nil {
		return nil, lxerrors.New("getting vm list from virtualbox", err)
	}
	vmNames := []string{}
	lines := strings.Split(string(out), "\n")
	r, err := regexp.Compile("\"(.*)\"")
	if err != nil {
		return nil, lxerrors.New("compiling regex", err)
	}
	for _, line := range lines {
		vmName := r.FindStringSubmatch(line)
		if len(vmName) > 0 {
			vmNames = append(vmNames, vmName[1])
		}
	}
	vms := []*VboxVm{}
	for _, vmName := range vmNames {
		if strings.Contains(vmName, "inaccessible") {
			continue
		}
		logrus.Debugf("found vm: " + vmName)
		vmInfo, err := vboxManage("showvminfo", vmName)
		if err != nil {
			return nil, lxerrors.New("getting vm info for "+vmName, err)
		}
		vm, err := parseVmInfo(string(vmInfo))
		if err != nil {
			return nil, lxerrors.New("parsing vm info string", err)
		}
		vm.Name = vmName
		vms = append(vms, vm)
	}

	return vms, nil
}

func GetVm(vmNameOrId string) (*VboxVm, error) {
	vms, err := Vms()
	if err != nil {
		return nil, lxerrors.New("getting vm list", err)
	}
	for _, vm := range vms {
		if vm.Name == vmNameOrId {
			return vm, nil
		}
	}
	return nil, lxerrors.New("vm "+ vmNameOrId +" not found", err)
}

func CreateVm(vmName, baseFolder, adapterName string, adapterType config.VirtualboxAdapterType) error {
	var nicArgs []string
	switch adapterType {
	case config.BridgedAdapter:
		nicArgs = []string{"modifyvm", vmName, "--nic1", "bridged", "--bridgeadapter1", adapterName, "--nictype1", "virtio"}
	case config.HostOnlyAdapter:
		nicArgs = []string{"modifyvm", vmName, "--nic1", "hostonly", "--hostonlyadapter1", adapterName, "--nictype1", "virtio"}
	default:
		return lxerrors.New(string(adapterType)+" not a valid adapter type, must specify either "+string(config.BridgedAdapter)+" or "+string(config.HostOnlyAdapter)+" network config", nil)
	}
	if _, err := vboxManage("createvm", "--name", vmName, "--basefolder", baseFolder, "-ostype", "Linux26_64"); err != nil {
		return lxerrors.New("creating vm", err)
	}
	if _, err := vboxManage("registervm", filepath.Join(baseFolder, vmName, fmt.Sprintf("%s.vbox", vmName))); err != nil {
		return lxerrors.New("registering vm", err)
	}
	if _, err := vboxManage("storagectl", vmName, "--name", "SCSI", "--add", "scsi", "--controller", "LsiLogic"); err != nil {
		return lxerrors.New("adding scsi storage controller", err)
	}
	//NIC ORDER MATTERS
	if _, err := vboxManage(nicArgs...); err != nil {
		return lxerrors.New("setting "+string(adapterType)+" networking on vm", err)
	}
	if _, err := vboxManage("modifyvm", vmName, "--nic2", "nat", "--nictype2", "virtio"); err != nil {
		return lxerrors.New("setting nat networking on vm", err)
	}
	return nil
}

func CreateVmNatless(vmName, baseFolder, adapterName string, adapterType config.VirtualboxAdapterType) error {
	var nicArgs []string
	switch adapterType {
	case config.BridgedAdapter:
		nicArgs = []string{"modifyvm", vmName, "--nic1", "bridged", "--bridgeadapter1", adapterName, "--nictype1", "virtio"}
	case config.HostOnlyAdapter:
		nicArgs = []string{"modifyvm", vmName, "--nic1", "hostonly", "--hostonlyadapter1", adapterName, "--nictype1", "virtio"}
	default:
		return lxerrors.New(string(adapterType)+" not a valid adapter type, must specify either "+string(config.BridgedAdapter)+" or "+string(config.HostOnlyAdapter)+" network config", nil)
	}
	if _, err := vboxManage("createvm", "--name", vmName, "--basefolder", baseFolder, "-ostype", "Linux26_64"); err != nil {
		return lxerrors.New("creating vm", err)
	}
	if _, err := vboxManage("registervm", filepath.Join(baseFolder, vmName, fmt.Sprintf("%s.vbox", vmName))); err != nil {
		return lxerrors.New("registering vm", err)
	}
	if _, err := vboxManage("storagectl", vmName, "--name", "SCSI", "--add", "scsi", "--controller", "LsiLogic"); err != nil {
		return lxerrors.New("adding scsi storage controller", err)
	}
	if _, err := vboxManage(nicArgs...); err != nil {
		return lxerrors.New("setting "+string(adapterType)+" networking on vm", err)
	}
	return nil
}

func ConfigureVmNetwork(vmName, adapterName string, adapterType config.VirtualboxAdapterType) error {
	var nicArgs []string
	switch adapterType {
	case config.BridgedAdapter:
		nicArgs = []string{"modifyvm", vmName, "--nic1", "bridged", "--bridgeadapter1", adapterName, "--nictype1", "virtio"}
	case config.HostOnlyAdapter:
		nicArgs = []string{"modifyvm", vmName, "--nic1", "hostonly", "--hostonlyadapter1", adapterName, "--nictype1", "virtio"}
	default:
		return lxerrors.New(string(adapterType)+" not a valid adapter type, must specify either "+string(config.BridgedAdapter)+" or "+string(config.HostOnlyAdapter)+" network config", nil)
	}
	if _, err := vboxManage(nicArgs...); err != nil {
		return lxerrors.New("setting "+string(adapterType)+" networking on vm", err)
	}
	return nil
}

func DestroyVm(vmNameOrId string) error {
	if _, err := vboxManage("unregistervm", vmNameOrId, "--delete"); err != nil {
		return lxerrors.New("unregistering and deleting vm", err)
	}
	return nil
}

func PowerOnVm(vmNameOrId string) error {
	_, err := vboxManage("startvm", vmNameOrId, "--type", "headless")
	return err
}

func PowerOffVm(vmNameOrId string) error {
	_, err := vboxManage("controlvm", vmNameOrId, "poweroff")
	return err
}

func AttachDisk(vmNameOrId, vmdkPath string, controllerPort int) error {
	if _, err := vboxManage("storageattach", vmNameOrId, "--storagectl", "SCSI", "--port", fmt.Sprintf("%v", controllerPort), "--type", "hdd", "--medium", vmdkPath); err != nil {
		return lxerrors.New("attaching storage", err)
	}
	return nil
}

func DetachDisk(vmNameOrId string, controllerPort int) error {
	if _, err := vboxManage("storageattach", vmNameOrId, "--storagectl", "SCSI", "--port", fmt.Sprintf("%v", controllerPort), "--type", "hdd", "--medium", "none"); err != nil {
		return lxerrors.New("attaching storage", err)
	}
	return nil
}

func formatMac(rawMac string) string {
	return strings.ToLower(rawMac[0:2] + ":" + rawMac[2:4] + ":" + rawMac[4:6] + ":" + rawMac[6:8] + ":" + rawMac[8:10] + ":" + rawMac[10:12])
}
