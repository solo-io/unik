package vsphereclient

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	uniklog "github.com/emc-advanced-dev/unik/pkg/util/log"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	vspheretypes "github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

type VsphereClient struct {
	u  *url.URL
	ds string
}

func NewVsphereClient(u *url.URL, datastore string) *VsphereClient {
	return &VsphereClient{
		u:  u,
		ds: datastore,
	}
}

func (vc *VsphereClient) newGovmomiClient() (*govmomi.Client, error) {
	c, err := govmomi.NewClient(context.TODO(), vc.u, true)
	if err != nil {
		return nil, lxerrors.New("creating new govmovi client", err)
	}
	return c, nil
}

func (vc *VsphereClient) newGovmomiFinder() (*find.Finder, error) {
	c, err := vc.newGovmomiClient()
	if err != nil {
		return nil, err
	}
	f := find.NewFinder(c.Client, true)

	// Find one and only datacenter
	dc, err := f.DefaultDatacenter(context.TODO())
	if err != nil {
		return nil, lxerrors.New("finding default datacenter", err)
	}

	// Make future calls local to this datacenter
	f.SetDatacenter(dc)
	return f, nil
}

func (vc *VsphereClient) Vms() ([]mo.VirtualMachine, error) {
	f, err := vc.newGovmomiFinder()
	if err != nil {
		return nil, lxerrors.New("creating new govmomi finder", err)
	}
	vms, err := f.VirtualMachineList(context.TODO(), "*")
	if err != nil {
		return nil, lxerrors.New("retrieving virtual machine list from finder", err)
	}
	vmList := []mo.VirtualMachine{}
	for _, vm := range vms {
		managedVms := []mo.VirtualMachine{}
		pc := property.DefaultCollector(vm.Client())
		refs := make([]vspheretypes.ManagedObjectReference, 0, len(vms))
		refs = append(refs, vm.Reference())
		err = pc.Retrieve(context.TODO(), refs, nil, &managedVms)
		if err != nil {
			return nil, lxerrors.New("retrieving managed vms property of vm "+vm.String(), err)
		}
		if len(managedVms) < 1 {
			return nil, lxerrors.New("0 managed vms found for vm "+vm.String(), nil)
		}
		logrus.WithFields(logrus.Fields{"vm": managedVms[0]}).Debugf("read vm from govmomi client")
		vmList = append(vmList, managedVms[0])
	}
	return vmList, nil
}

func (vc *VsphereClient) GetVm(vmName string) (mo.VirtualMachine, error) {
	vms, err := vc.Vms()
	if err != nil {
		return mo.VirtualMachine{}, lxerrors.New("getting vsphere vm list", err)
	}
	for _, vm := range vms {
		if vm.Name == vmName {
			return vm, nil
		}
	}
	return mo.VirtualMachine{}, lxerrors.New("no vm found with name "+vmName, nil)
}

func (vc *VsphereClient) GetVmIp(vmName string) (string, error) {
	vm, err := vc.GetVm(vmName)
	if err != nil {
		return "", lxerrors.New("getting vsphere vm", err)
	}
	if vm.Guest == nil {
		return "", lxerrors.New("vm has no guest info", nil)
	}
	return vm.Guest.IpAddress, nil
}

func (vc *VsphereClient) CreateVm(vmName string, memoryMb int) error {
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"govc",
		"vm.create",
		"-k",
		"-u", formatUrl(vc.u),
		"--force=true",
		fmt.Sprintf("--m=%v", memoryMb),
		"--on=false",
		vmName,
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc vm.create "+vmName, err)
	}
	return nil
}

func (vc *VsphereClient) DestroyVm(vmName string) error {
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"govc",
		"vm.destroy",
		"-k",
		"-u", formatUrl(vc.u),
		vmName,
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc vm.destroy "+vmName, err)
	}
	return nil
}

func (vc *VsphereClient) Mkdir(folder string) error {
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"govc",
		"datastore.mkdir",
		"-k",
		"-u", formatUrl(vc.u),
		folder,
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc datastore.mkdir "+folder, err)
	}
	return nil
}

func (vc *VsphereClient) Rmdir(folder string) error {
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"govc",
		"datastore.rm",
		"-k",
		"-u", formatUrl(vc.u),
		folder,
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc datastore.rm "+folder, err)
	}
	return nil
}

func (vc *VsphereClient) ImportVmdk(vmdkPath, folder string) error {
	vmdkFolder := filepath.Dir(vmdkPath)
	cmd := exec.Command("docker", "run", "--rm", "-v", vmdkFolder+":"+vmdkFolder,
		"vsphere-client",
		"govc",
		"import.vmdk",
		"-k",
		"-u", formatUrl(vc.u),
		vmdkPath,
		folder,
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc import.vmdk "+folder, err)
	}
	return nil
}

func (vc *VsphereClient) UploadFile(srcFile, dest string) error {
	srcDir := filepath.Dir(srcFile)
	cmd := exec.Command("docker", "run", "--rm", "-v", srcDir+":"+srcDir,
		"vsphere-client",
		"govc",
		"datastore.upload",
		"-k",
		"-u", formatUrl(vc.u),
		srcFile,
		dest,
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc datastore.upload", err)
	}
	return nil
}

func (vc *VsphereClient) DownloadFile(remoteFile, localFile string) error {
	localDir := filepath.Dir(localFile)
	cmd := exec.Command("docker", "run", "--rm", "-v", localDir+":"+localDir,
		"vsphere-client",
		"govc",
		"datastore.download",
		"-k",
		"-u", formatUrl(vc.u),
		remoteFile,
		localFile,
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc datastore.upload", err)
	}
	return nil
}

func (vc *VsphereClient) CopyVmdk(src, dest string) error {
	password, _ := vc.u.User.Password()
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"java",
		"-jar",
		"/vsphere-client.jar",
		"CopyVirtualDisk",
		vc.u.String(),
		vc.u.User.Username(),
		password,
		"["+vc.ds+"] "+src,
		"["+vc.ds+"] "+dest,
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running vsphere-client.jar CopyVirtualDisk "+src+" "+dest, err)
	}
	return nil
}

func (vc *VsphereClient) Ls(dir string) ([]string, error) {
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"govc",
		"datastore.ls",
		"-k",
		"-u", formatUrl(vc.u),
		dir,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, lxerrors.New("failed running govc datastore.ls "+dir, err)
	}
	split := strings.Split(string(out), "\n")
	contents := []string{}
	for _, content := range split {
		if content != "" {
			contents = append(contents, content)
		}
	}
	return contents, nil
}

func (vc *VsphereClient) PowerOnVm(vmName string) error {
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"govc",
		"vm.power",
		"--on=true",
		"-k",
		"-u", formatUrl(vc.u),
		vmName,
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc vm.power (on)", err)
	}
	return nil
}

func (vc *VsphereClient) PowerOffVm(vmName string) error {
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"govc",
		"vm.power",
		"--off=true",
		"-k",
		"-u", formatUrl(vc.u),
		vmName,
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc vm.power (off)", err)
	}
	return nil
}

func (vc *VsphereClient) AttachDisk(vmName, vmdkPath string, controllerKey int) error {
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
		"["+vc.ds+"] "+vmdkPath,
		fmt.Sprintf("%v", controllerKey),
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running vsphere-client.jar AttachVmdk", err)
	}
	return nil
}

func (vc *VsphereClient) DetachDisk(vmName string, controllerKey int) error {
	password, _ := vc.u.User.Password()
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"java",
		"-jar",
		"/vsphere-client.jar",
		"VmDetachDisk",
		vc.u.String(),
		vc.u.User.Username(),
		password,
		vmName,
		fmt.Sprintf("%v", controllerKey),
	)
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running vsphere-client.jar DetachVmdk", err)
	}
	return nil
}

func formatUrl(u *url.URL) string {
	return "https://" + strings.TrimPrefix(strings.TrimPrefix(u.String(), "http://"), "https://")
}
