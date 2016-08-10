package vsphereclient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"golang.org/x/net/context"
)

type VsphereClient struct {
	u  *url.URL
	ds string
	dc string
}

func NewVsphereClient(u *url.URL, datastore, datacenter string) *VsphereClient {
	return &VsphereClient{
		u:  u,
		ds: datastore,
		dc: datacenter,
	}
}

func (vc *VsphereClient) newGovmomiClient() (*govmomi.Client, error) {
	c, err := govmomi.NewClient(context.TODO(), vc.u, true)
	if err != nil {
		return nil, errors.New("creating new govmovi client", err)
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
		return nil, errors.New("finding default datacenter", err)
	}

	// Make future calls local to this datacenter
	f.SetDatacenter(dc)
	return f, nil
}

func (vc *VsphereClient) GetVmByUuid(uuid string) (*VirtualMachine, error) {

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"govc",
		"vm.info",
		"-k",
		"-u", formatUrl(vc.u),
		"--json",
		"--vm.uuid=" + uuid,
	}
	logrus.WithField("command", args).Debugf("running command")
	out, err := container.CombinedOutput(args...)
	if err != nil {
		return nil, errors.New("failed running govc vm.info "+uuid, err)
	}
	var vm VmInfo
	if err := json.Unmarshal(out, &vm); err != nil {
		return nil, errors.New("unmarshalling json: "+string(out), err)
	}
	if len(vm.VirtualMachines) < 1 {
		return nil, errors.New("returned virtualmachines had len 0; does vm exist? "+string(out), nil)
	}
	return &vm.VirtualMachines[0], nil
}

func (vc *VsphereClient) GetVm(name string) (*VirtualMachine, error) {

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"govc",
		"vm.info",
		"-k",
		"-u", formatUrl(vc.u),
		"--json",
		name,
	}
	logrus.WithField("command", args).Debugf("running command")
	out, err := container.CombinedOutput(args...)
	if err != nil {
		return nil, errors.New("failed running govc vm.info "+name, err)
	}
	var vm VmInfo
	if err := json.Unmarshal(out, &vm); err != nil {
		return nil, errors.New("unmarshalling json: "+string(out), err)
	}
	if len(vm.VirtualMachines) < 1 {
		return nil, errors.New("returned virtualmachines had len 0; does vm exist? "+string(out), nil)
	}
	return &vm.VirtualMachines[0], nil
}

func (vc *VsphereClient) GetVmIp(vmName string) (string, error) {
	vm, err := vc.GetVm(vmName)
	if err != nil {
		return "", errors.New("getting vsphere vm", err)
	}
	return vm.Guest.IPAddress, nil
}

func (vc *VsphereClient) CreateVm(vmName string, memoryMb int, networkType types.VsphereNetworkType, networkLabel string) error {

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"govc",
		"vm.create",
		"-k",
		"-u", formatUrl(vc.u),
		"--force=true",
		fmt.Sprintf("--m=%v", memoryMb),
		"--on=false",
		"-ds", vc.ds,
		fmt.Sprintf("-net.adapter=%s", networkType),
	}
	if networkLabel != "" {
		args = append(args, fmt.Sprintf("-net=%s", networkLabel))
	}
	args = append(args, vmName)

	if err := container.Run(args...); err != nil {
		return errors.New("failed running govc vm.create "+vmName, err)
	}
	return nil
}

func (vc *VsphereClient) DestroyVm(vmName string) error {

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"govc",
		"vm.destroy",
		"-k",
		"-u", formatUrl(vc.u),
		vmName,
	}
	if err := container.Run(args...); err != nil {
		return errors.New("failed running govc vm.destroy "+vmName, err)
	}
	return nil
}

func (vc *VsphereClient) Mkdir(folder string) error {

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"govc",
		"datastore.mkdir",
		"-ds", vc.ds,
		"-k",
		"-u", formatUrl(vc.u),
		folder,
	}
	if err := container.Run(args...); err != nil {
		logrus.WithError(err).Warnf("failed running govc datastore.mkdir " + folder)
	}
	return nil
}

func (vc *VsphereClient) Rmdir(folder string) error {

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"govc",
		"datastore.rm",
		"-ds", vc.ds,
		"-k",
		"-u", formatUrl(vc.u),
		folder,
	}
	if err := container.Run(args...); err != nil {
		return errors.New("failed running govc datastore.rm "+folder, err)
	}
	return nil
}

func (vc *VsphereClient) ImportVmdk(vmdkPath, remoteFolder string) error {
	vmdkFolder, err := filepath.Abs(filepath.Dir(vmdkPath))
	if err != nil {
		return errors.New("getting aboslute path for "+vmdkFolder, err)
	}

	container := unikutil.NewContainer("vsphere-client").WithVolume(vmdkFolder, vmdkFolder)
	args := []string{
		"govc",
		"import.vmdk",
		"-ds", vc.ds,
		"-k",
		"-u", formatUrl(vc.u),
		vmdkPath,
		remoteFolder,
	}
	if err := container.Run(args...); err != nil {
		return errors.New("failed running govc import.vmdk "+remoteFolder, err)
	}
	return nil
}

func (vc *VsphereClient) UploadFile(srcFile, dest string) error {
	srcDir := filepath.Dir(srcFile)

	container := unikutil.NewContainer("vsphere-client").WithVolume(srcDir, srcDir)
	args := []string{
		"govc",
		"datastore.upload",
		"-ds", vc.ds,
		"-k",
		"-u", formatUrl(vc.u),
		srcFile,
		dest,
	}
	if err := container.Run(args...); err != nil {
		return errors.New("failed running govc datastore.upload", err)
	}
	return nil
}

func (vc *VsphereClient) DownloadFile(remoteFile, localFile string) error {
	localDir := filepath.Dir(localFile)

	container := unikutil.NewContainer("vsphere-client").WithVolume(localDir, localDir)
	args := []string{
		"govc",
		"datastore.download",
		"-ds", vc.ds,
		"-k",
		"-u", formatUrl(vc.u),
		remoteFile,
		localFile,
	}
	if err := container.Run(args...); err != nil {
		return errors.New("failed running govc datastore.upload", err)
	}
	return nil
}

func (vc *VsphereClient) CopyVmdk(src, dest string) error {
	password, _ := vc.u.User.Password()

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"java",
		"-jar",
		"/vsphere-client.jar",
		"CopyVirtualDisk",
		vc.u.String(),
		vc.u.User.Username(),
		password,
		vc.dc,
		"[" + vc.ds + "] " + src,
		"[" + vc.ds + "] " + dest,
	}
	if err := container.Run(args...); err != nil {
		return errors.New("failed running vsphere-client.jar CopyVirtualDisk "+src+" "+dest, err)
	}
	return nil
}

func (vc *VsphereClient) CopyFile(src, dest string) error {
	password, _ := vc.u.User.Password()
	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"java",
		"-jar",
		"/vsphere-client.jar",
		"CopyFile",
		vc.u.String(),
		vc.u.User.Username(),
		password,
		vc.dc,
		"[" + vc.ds + "] " + src,
		"[" + vc.ds + "] " + dest,
	}
	if err := container.Run(args...); err != nil {
		lastSlash := strings.LastIndex(dest, "/")
		directory := "/"
		file := dest[lastSlash+1:]
		if lastSlash != -1 {
			directory += dest[0:lastSlash]
			file = dest[lastSlash+1:]
		}
		files, err := vc.Ls(directory)
		if err != nil {
			return errors.New("failed running vsphere-client.jar CopyFile "+src+" "+dest, err)
		}
		if !unikutil.StringInSlice(file, files) {
			return errors.New("failed running vsphere-client.jar CopyFile "+src+" "+dest, err)
		}
	}
	return nil
}

func (vc *VsphereClient) Ls(dir string) ([]string, error) {

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"govc",
		"datastore.ls",
		"-ds", vc.ds,
		"-k",
		"-u", formatUrl(vc.u),
		dir,
	}
	out, err := container.Output(args...)
	if err != nil {
		return nil, errors.New("failed running govc datastore.ls "+dir, err)
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

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"govc",
		"vm.power",
		"--on=true",
		"-k",
		"-u", formatUrl(vc.u),
		vmName,
	}
	if err := container.Run(args...); err != nil {
		return errors.New("failed running govc vm.power (on)", err)
	}
	return nil
}

func (vc *VsphereClient) PowerOffVm(vmName string) error {

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"govc",
		"vm.power",
		"--off=true",
		"-k",
		"-u", formatUrl(vc.u),
		vmName,
	}
	if err := container.Run(args...); err != nil {
		return errors.New("failed running govc vm.power (off)", err)
	}
	return nil
}

func (vc *VsphereClient) AttachDisk(vmName, vmdkPath string, controllerKey int, deviceType types.StorageDriver) error {
	password, _ := vc.u.User.Password()

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"java",
		"-jar",
		"/vsphere-client.jar",
		"VmAttachDisk",
		vc.u.String(),
		vc.u.User.Username(),
		password,
		vmName,
		"[" + vc.ds + "] " + vmdkPath,
		string(deviceType),
		fmt.Sprintf("%v", controllerKey),
	}
	if err := container.Run(args...); err != nil {
		return errors.New("failed running vsphere-client.jar AttachVmdk", err)
	}
	return nil
}

func (vc *VsphereClient) DetachDisk(vmName string, controllerKey int, deviceType types.StorageDriver) error {
	password, _ := vc.u.User.Password()

	container := unikutil.NewContainer("vsphere-client")
	args := []string{
		"java",
		"-jar",
		"/vsphere-client.jar",
		"VmDetachDisk",
		vc.u.String(),
		vc.u.User.Username(),
		password,
		vmName,
		string(deviceType),
		fmt.Sprintf("%v", controllerKey),
	}
	if err := container.Run(args...); err != nil {
		return errors.New("failed running vsphere-client.jar DetachVmdk", err)
	}
	return nil
}

func formatUrl(u *url.URL) string {
	return "https://" + strings.TrimPrefix(strings.TrimPrefix(u.String(), "http://"), "https://")
}
