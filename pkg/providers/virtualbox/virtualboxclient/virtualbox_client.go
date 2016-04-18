package virtualboxclient

import (
	"github.com/layer-x/layerx-commons/lxerrors"
	"os/exec"
	"strings"
	"github.com/pwnall/vbox"
	"net/url"
	"fmt"
	"time"
	"github.com/Sirupsen/logrus"
	uniklog "github.com/emc-advanced-dev/unik/pkg/util/log"
)

func vboxManage(args ...string) (string, error) {
	cmd := exec.Command("VBoxManage", args...)
	logrus.WithField("command", cmd.Args).Debugf("running VBoxManage command")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", string(out))
	}
	return string(out), nil
}

func Vms() ([]vbox.Machine, error) {
	vms, err := vbox.GetMachines()
	if err != nil {
		return nil, lxerrors.New("getting vm list from virtualbox", err)
	}
	return vms, nil
}

func CreateVm(vmName string) error {
	vm, err := vbox.CreateMachine(vmName, "Other", "")
	if err != nil {
		return lxerrors.New("creating vm", err)
	}
	defer vm.Release()
	if err = vm.Register(); err != nil {
		return lxerrors.New("registering vm", err)
	}

	if _, err := vboxManage("storagectl", vmName, "--name", "SCSI", "--add", "scsi", "--controller", "LsiLogic"); err != nil {
		return lxerrors.New("adding scsi storage controller", err)
	}

	if _, err := vboxManage("storagectl", vmName, "--name", "SCSI", "--add", "scsi", "--controller", "LsiLogic"); err != nil {
		return lxerrors.New("adding scsi storage controller", err)
	}
	if _, err := vboxManage("modifyvm", vmName, "--nic1", "bridged", "--bridgeadapter1", c.Bridge, "--nictype1", "virtio"); err != nil {
		return lxerrors.New("setting bridged networking on vm", err)
	}
	return nil
}

func DestroyVm(vmName string) error {
	vm, err := vbox.FindMachine(vmName)
	if err != nil {
		return lxerrors.New("finding machine "+vmName, err)
	}
	defer vm.Release()
	media, err := vm.Unregister(vbox.CleanupMode_DetachAllReturnHardDisksOnly)
	if err != nil {
		return lxerrors.New("unregistering vm", err)
	}
	progress, err := vm.DeleteConfig(media)
	if err != nil {
		return lxerrors.New("deleting media", err)
	}
	if err = progress.WaitForCompletion(-1); err != nil; {
		return lxerrors.New("waiting for delete media config to complete", err)
	}
	percent, err := progress.GetPercent()
	if err != nil {
		return lxerrors.New("could not get progress percent", err)
	}
	logrus.Debugf("finished deleting media config for %s: %v percent", vmName, percent)

	if percent != 100 {
		return lxerrors.New("config deletion stopped at "+fmt.Sprintf("%v", percent), err)
	}
	result, err := progress.GetResultCode()
	if err != nil {
		return lxerrors.New("getting result code from config deletion", err)
	}
	if result != 0 {
		return lxerrors.New(fmt.Sprintf("config deletion failed with code %v", result), nil)
	}
	return nil
}

func PowerOnVm(vmName string) error {


	machine, err := vbox.FindMachine(vmName)
	if err != nil {
		return lxerrors.New("finding machine "+vmName, err)
	}
	defer machine.Release()

	session := vbox.Session{}
	if err := session.Init(); err != nil {
		return lxerrors.New("session init", err)
	}
	defer session.Release()

	progress, err := machine.Launch(session, "gui", "");
	if err != nil {
		return lxerrors.New("launching vm session", err)
	}
	defer progress.Release()
	defer func() {
		if err = session.UnlockMachine(); err != nil {
			logrus.WithError(err).Errorf("failed to unlock machine %s", vmName)
			return
		}
		for {
			state, err := session.GetState()
			if err != nil {
				logrus.WithError(err).Errorf("getting session state")
				return
			}
			logrus.Debugf("Session state: %s for vm %s", state, vmName)
			if state == vbox.SessionState_Unlocked {
				break
			}
			time.Sleep(300 * time.Millisecond)
		}
		// TODO(pwnall): Figure out how to get rid of this timeout. The VM should
		//     be unlocked, according to the check above, but unregistering the VM
		//     fails if we don't wait.
		time.Sleep(300 * time.Millisecond)
	}()


	if err = progress.WaitForCompletion(50000); err != nil {
		return lxerrors.New("launching vm session", err)
	}

	console, err := session.GetConsole()
	if err != nil {
		return lxerrors.New("getting vm console", err)
	}
	defer console.Release()

	console.PowerDown()

	return nil
}

func PowerOffVm(vmName string) error {
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"govc",
		"vm.power",
		"--off=true",
		"-k",
		"-u", formatUrl(vc.u),
		vmName,
	)
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc vm.power (off)", err)
	}
	return nil
}

func AttachVmdk(vmName, vmdkPath string) error {
	password, _ := vc.u.User.Password()
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"java",
		"-jar",
		"/vsphere-client.jar",
		"VmAttachDisk",
		vc.u.String(),
		vc.u.User.Username(),
		password,
		vmName,
		"[datastore1] " + vmdkPath,
		"200", //TODO: is this right?
	)
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running vsphere-client.jar command")
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running vsphere-client.jar AttachVmdk", err)
	}
	return nil
}

func formatUrl(u *url.URL) string {
	return "https://" + strings.TrimPrefix(strings.TrimPrefix(u.String(), "http://"), "https://")
}
