package api

import (
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"net/url"
	"github.com/layer-x/layerx-commons/lxerrors"
	"golang.org/x/net/context"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/property"
	vspheretypes "github.com/vmware/govmomi/vim25/types"
	"os/exec"
	"strings"
	"path/filepath"
	"github.com/Sirupsen/logrus"
	uniklog "github.com/emc-advanced-dev/unik/pkg/util/log"
)

type VsphereClient struct {
	u *url.URL
}

func NewVsphereClient(u *url.URL) (*VsphereClient, error) {
	return &VsphereClient{
		u: u,
	}, nil
}

func (c *VsphereClient) newGovmomiClient() (*govmomi.Client, error) {
	c, err := govmomi.NewClient(context.TODO(), c.u, true)
	if err != nil {
		return nil, lxerrors.New("creating new govmovi client", err)
	}
	return c, nil
}

func (c *VsphereClient) newGovmomiFinder() (*find.Finder, error) {
	c, err := c.newGovmomiClient()
	if err != nil {
		return err
	}
	f := find.NewFinder(c.Client, true)

	// Find one and only datacenter
	dc, err := f.DefaultDatacenter(context.TODO())
	if err != nil {
		return nil, lxerrors.New("finding default datacenter", err)
	}

	// Make future calls local to this datacenter
	f.SetDatacenter(dc)
	return f
}


func (vc *VsphereClient) Vms() ([]mo.VirtualMachine, error) {
	f, err := vc.newGovmomiFinder()
	if err != nil {
		return lxerrors.New("creating new govmomi finder", err)
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
			return nil, lxerrors.New("retrieving managed vms property of vm " + vm.String(), err)
		}
		if len(managedVms) < 1 {
			return nil, lxerrors.New("0 managed vms found for vm " + vm.String(), nil)
		}
		logrus.WithFields(logrus.Fields{"vm": managedVms[0]}).Debugf("read vm from govmomi client")
		vmList = append(vmList, managedVms[0])
	}
	return vmList, nil
}

func (vc *VsphereClient) CreateVm(vmName, annotation, memoryMb string) error {
	cmd := exec.Command("docker", "run", "--rm",
		"vsphere-client",
		"govc",
		"vm.create",
		"-k",
		"-u", formatUrl(vc.u),
		"--annotation=" + annotation,
		"--force=true",
		"--m="+memoryMb,
		"--on=false",
		vmName,
	)
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc vm.create " + vmName, err)
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
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc vm.destroy " + vmName, err)
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
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc datastore.mkdir " + folder, err)
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
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc datastore.rm " + folder, err)
	}
	return nil
}

func (vc *VsphereClient) ImportVmdk(vmdkPath, folder string) error {
	vmdkFolder := filepath.Dir(vmdkPath)
	cmd := exec.Command("docker", "run", "--rm", "-v", vmdkFolder + ":" + vmdkFolder,
		"vsphere-client",
		"govc",
		"import.vmdk",
		"-k",
		"-u", formatUrl(vc.u),
		vmdkPath,
		folder,
	)
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc import.vmdk " + folder, err)
	}
	return nil
}

func (vc *VsphereClient) UploadFile(srcFile, dest string) error {
	srcDir := filepath.Dir(srcFile)
	cmd := exec.Command("docker", "run", "--rm", "-v", srcDir + ":" + srcDir,
		"vsphere-client",
		"govc",
		"datastore.upload",
		"-k",
		"-u", formatUrl(vc.u),
		srcFile,
		dest,
	)
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running govc datastore.upload", err)
	}
	return nil
}

func (vc *VsphereClient) DownloadFile(remoteFile, localFile string) error {
	localDir := filepath.Dir(localFile)
	cmd := exec.Command("docker", "run", "--rm", "-v", localDir + ":" + localDir,
		"vsphere-client",
		"govc",
		"datastore.download",
		"-k",
		"-u", formatUrl(vc.u),
		remoteFile,
		localFile,
	)
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
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
		"[datastore1] " + src,
		"[datastore1] " + dest,
	)
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running vsphere-client.jar command")
	uniklog.LogCommand(cmd, true)
	err := cmd.Run()
	if err != nil {
		return lxerrors.New("failed running vsphere-client.jar CopyVirtualDisk " + src + " " + dest, err)
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
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
	out, err := cmd.Output()
	if err != nil {
		return nil, lxerrors.New("failed running govc datastore.ls " + dir, err)
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
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running govc command")
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

func (vc *VsphereClient) AttachVmdk(vmName, vmdkPath string) error {
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
