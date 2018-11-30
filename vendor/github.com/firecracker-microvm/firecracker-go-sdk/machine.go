// Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package firecracker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	ops "github.com/firecracker-microvm/firecracker-go-sdk/client/operations"
	log "github.com/sirupsen/logrus"
)

const (
	userAgent       = "firecracker-go-sdk"
	fcExecutable    = "./firecracker"
	terminalProgram = "xterm"
	// ConsoleXterm indicates that the machine's console should be presented in an xterm
	ConsoleXterm = "xterm"
	// ConsoleStdio indicates that the machine's console should re-use the parent's stdio streams
	ConsoleStdio = "stdio"
	// ConsoleNone indicates that the machine's console IO should be discarded
	ConsoleNone = "none"
)

// CPUTemplate defines a set of CPU features that are exposed by Firecracker
type CPUTemplate = models.CPUTemplate

// CPUTemplates known by Firecracker. These are passed through directly from the model.
const (
	CPUTemplateT2 = models.CPUTemplateT2
	CPUTemplateC3 = models.CPUTemplateC3
)

type Firecracker interface {
	PutLogger(ctx context.Context, logger *models.Logger) (*ops.PutLoggerNoContent, error)
	PutMachineConfiguration(ctx context.Context, cfg *models.MachineConfiguration) (*ops.PutMachineConfigurationNoContent, error)
	PutGuestBootSource(ctx context.Context, source *models.BootSource) (*ops.PutGuestBootSourceNoContent, error)
	PutGuestNetworkInterfaceByID(ctx context.Context, ifaceID string, ifaceCfg *models.NetworkInterface) (*ops.PutGuestNetworkInterfaceByIDNoContent, error)
	PutGuestDriveByID(ctx context.Context, driveID string, drive *models.Drive) (*ops.PutGuestDriveByIDNoContent, error)
	PutGuestVsockByID(ctx context.Context, vsockID string, vsock *models.Vsock) (*ops.PutGuestVsockByIDCreated, *ops.PutGuestVsockByIDNoContent, error)
	CreateSyncAction(ctx context.Context, info *models.InstanceActionInfo) (*ops.CreateSyncActionNoContent, error)
	PutMmds(ctx context.Context, metadata interface{}) (*ops.PutMmdsCreated, *ops.PutMmdsNoContent, error)
	GetMachineConfig() (*ops.GetMachineConfigOK, error)
}

// Config is a collection of user-configurable VMM settings
type Config struct {
	BinPath           string
	SocketPath        string
	LogFifo           string
	LogLevel          string
	MetricsFifo       string
	KernelImagePath   string
	KernelArgs        string
	CPUCount          int64
	HtEnabled         bool
	CPUTemplate       CPUTemplate
	MemInMiB          int64
	RootDrive         BlockDevice
	RootPartitionUUID string
	AdditionalDrives  []BlockDevice
	NetworkInterfaces []NetworkInterface
	VsockDevices      []VsockDevice
	Console           string
	Debug             bool
	machineCfg        models.MachineConfiguration

	// Allows for easier mock testing
	disableValidation bool
}

// Validate will ensure that the required fields are set and that
// the fields are valid values.
func (cfg *Config) Validate() error {
	if cfg.disableValidation {
		return nil
	}

	if _, err := os.Stat(cfg.KernelImagePath); err != nil {
		return fmt.Errorf("Failed to stat kernal image path, %q: %v", cfg.KernelImagePath, err)
	}
	if _, err := os.Stat(cfg.RootDrive.HostPath); err != nil {
		return fmt.Errorf("Failed to stat host path, %q: %v", cfg.RootDrive.HostPath, err)
	}

	// Check the non-existence of some files:
	if _, err := os.Stat(cfg.SocketPath); err == nil {
		return fmt.Errorf("Socket %s already exists", cfg.SocketPath)
	}

	return nil
}

// Machine is the main object for manipulating firecracker VMs
type Machine struct {
	binPath       string
	cfg           Config
	client        Firecracker
	cmd           *exec.Cmd
	logger        *log.Entry
	machineConfig models.MachineConfiguration // The actual machine config as reported by Firecracker
}

// Logger returns a logrus logger appropriate for logging hypervisor messages
func (m *Machine) Logger() *log.Entry {
	// return m.logger.WithField("subsystem", "firecracker-go-sdk")
	return m.logger
}

// NetworkInterface represents a Firecracker VMs network interface.
type NetworkInterface struct {
	MacAddress  string
	HostDevName string
	AllowMDDS   bool
}

// BlockDevice represents a host block device mapped to the firecracker VM.
// mode is either "ro" or "rw"
type BlockDevice struct {
	HostPath string
	Mode     string
}

// VsockDevice represents a vsock connection between the
// host and the guest VM.
type VsockDevice struct {
	Path string
	CID  uint32
}

// SocketPath returns the filesystem path to the socket used for VMM communication
func (m Machine) socketPath() string {
	return m.cfg.SocketPath
}

// LogFile returns the filesystem path of the VMM log
func (m Machine) LogFile() string {
	return m.cfg.LogFifo
}

// LogLevel returns the VMM log level.
func (m Machine) LogLevel() string {
	return m.cfg.LogLevel
}

// NewMachine initializes a new Machine instance
func NewMachine(cfg Config, opts ...Opt) (*Machine, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	m := &Machine{}

	for _, opt := range opts {
		opt(m)
	}

	if m.logger == nil {
		logger := log.New()

		if cfg.Debug {
			logger.SetLevel(log.DebugLevel)
		}

		m.logger = log.NewEntry(logger)
	}

	if m.client == nil {
		m.client = NewFirecrackerClient(cfg.SocketPath, m.logger, cfg.Debug)
	}

	m.logger.Debug("Called NewMachine()")

	m.cfg = cfg
	m.cfg.machineCfg = models.MachineConfiguration{
		VcpuCount:   cfg.CPUCount,
		MemSizeMib:  cfg.MemInMiB,
		HtEnabled:   cfg.HtEnabled,
		CPUTemplate: models.CPUTemplate(cfg.CPUTemplate),
	}

	if len(cfg.BinPath) > 0 {
		m.binPath = cfg.BinPath
	} else {
		m.binPath = fcExecutable
	}

	return m, nil
}

// Init starts the VMM and attaches drives and network interfaces.
func (m *Machine) Init(ctx context.Context) (chan error, error) {
	m.logger.Debug("Called Machine.Init()")
	var err error
	errchan, err := m.startVMM(ctx)
	if err != nil {
		return errchan, err
	}
	err = m.setupLogging(ctx)
	if err != nil {
		m.logger.Warnf("setupLogging() returned %s. Continuing anyway.", err)
	} else {
		m.logger.Debugf("back from setupLogging")
	}
	err = m.createMachine(ctx)
	if err != nil {
		m.stopVMM()
		return errchan, err
	}
	m.logger.Debug("createMachine returned")
	err = m.createBootSource(ctx, m.cfg.KernelImagePath, m.cfg.KernelArgs)
	if err != nil {
		m.stopVMM()
		return errchan, err
	}
	m.logger.Debug("createBootSource returned")

	err = m.attachDrive(ctx, m.cfg.RootDrive, 1, true)
	if err != nil {
		m.stopVMM()
		return errchan, err
	}
	m.logger.Debug("Root drive attachment complete")

	for id, dev := range m.cfg.AdditionalDrives {
		// id must be increased by 2 because firecracker uses 1-indexed arrays and the root drive occupies position 1.
		err = m.attachDrive(ctx, dev, id+2, false)
		if err != nil {
			m.logger.Errorf("While attaching secondary drive %s, got error %s", dev.HostPath, err)
			m.stopVMM()
			return errchan, err
		}
		m.logger.Debugf("attachDrive returned for %s", dev.HostPath)
	}
	for id, iface := range m.cfg.NetworkInterfaces {
		err = m.createNetworkInterface(ctx, iface, id+1)
		if err != nil {
			m.stopVMM()
			return errchan, err
		}
		m.logger.Debugf("createNetworkInterface returned for %s", iface.HostDevName)
	}
	for _, dev := range m.cfg.VsockDevices {
		err = m.addVsock(ctx, dev)
		if err != nil {
			m.stopVMM()
			return errchan, err
		}
	}

	m.logger.Debugf("returning from Machine.Init(), RootDrive=%s", m.cfg.RootDrive.HostPath)
	return errchan, nil
}

// startVMM starts the firecracker vmm process and configures logging.
func (m *Machine) startVMM(ctx context.Context) (chan error, error) {
	var cmd *exec.Cmd
	var exitchannel = make(chan error)
	m.logger.Printf("Called startVMM(), setting up a VMM on %s", m.cfg.SocketPath)
	switch m.cfg.Console {
	case ConsoleXterm:
		cmd = exec.CommandContext(ctx, terminalProgram, "-e", m.binPath, "--api-sock", m.cfg.SocketPath)
	case ConsoleStdio:
		cmd = exec.CommandContext(ctx, m.binPath, "--api-sock", m.cfg.SocketPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	default:
		cmd = exec.CommandContext(ctx, m.binPath, "--api-sock", m.cfg.SocketPath)
	}
	m.cmd = cmd

	err := cmd.Start()
	if err != nil {
		m.logger.Errorf("Failed to start VMM: %s", err)
		return exitchannel, err
	}
	m.logger.Debugf("VMM started, PID=%d, socket path is %s", cmd.Process.Pid, m.cfg.SocketPath)

	// Set up an exit handler
	go func() {
		err := cmd.Wait()
		if err != nil {
			m.logger.Warnf("firecracker exited: %s", err.Error())
		} else {
			m.logger.Printf("firecracker exited: status=0")
		}
		os.Remove(m.cfg.SocketPath)
		os.Remove(m.cfg.LogFifo)
		os.Remove(m.cfg.MetricsFifo)
		exitchannel <- err
	}()

	// Set up a signal hander and pass INT, QUIT, and TERM through to firecracker
	vmchan := make(chan error)
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, os.Interrupt,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGABRT)
	m.logger.Debugf("Setting up signal handler")
	go func() {
		select {
		case sig := <-sigchan:
			m.logger.Printf("Caught signal %s", sig)
			cmd.Process.Signal(sig)
		case err = <-vmchan:
			exitchannel <- err
		}
	}()

	// Wait for firecracker to initialize:
	err = m.waitForSocket(3*time.Second, exitchannel)
	if err != nil {
		msg := fmt.Sprintf("Firecracker did not create API socket %s: %s", m.cfg.SocketPath, err)
		err = errors.New(msg)
		return exitchannel, err
	}

	m.logger.Debugf("returning from startVMM()")
	return exitchannel, nil
}

//StopVMM stops the current VMM.
func (m *Machine) StopVMM() error {
	return m.stopVMM()
}

func (m *Machine) stopVMM() error {
	if m.cmd != nil {
		log.Debug("stopVMM(): sending sigterm to firecracker")
		return m.cmd.Process.Signal(syscall.SIGTERM)
	}
	log.Debug("stopVMM(): no firecracker process running, not sending a signal")

	// don't return an error if the process isn't even running
	return nil
}

// createFifos sets up the firecracker logging and metrics fifos
func createFifos(logFifo, metricsFifo string) error {
	log.Debugf("Creating FIFO %s", logFifo)
	err := syscall.Mkfifo(logFifo, 0700)
	if err != nil {
		return err
	}
	log.Debugf("Creating FIFO %s", metricsFifo)
	err = syscall.Mkfifo(metricsFifo, 0700)
	return err
}

func (m *Machine) setupLogging(ctx context.Context) error {
	if len(m.cfg.LogFifo) == 0 || len(m.cfg.MetricsFifo) == 0 {
		// No logging configured
		m.logger.Printf("VMM logging and metrics disabled.")
		return nil
	}

	err := createFifos(m.cfg.LogFifo, m.cfg.MetricsFifo)
	if err != nil {
		m.logger.Errorf("Unable to set up logging: %s", err)
		return err
	}

	m.logger.Debug("Created metrics and logging fifos.")

	l := models.Logger{
		LogFifo:       m.cfg.LogFifo,
		Level:         m.cfg.LogLevel,
		MetricsFifo:   m.cfg.MetricsFifo,
		ShowLevel:     true,
		ShowLogOrigin: false,
	}

	resp, err := m.client.PutLogger(ctx, &l)
	if err == nil {
		m.logger.Printf("Configured VMM logging to %s, metrics to %s: %s",
			m.cfg.LogFifo, m.cfg.MetricsFifo, resp.Error())
	}
	return err
}

func (m *Machine) createMachine(ctx context.Context) error {
	resp, err := m.client.PutMachineConfiguration(ctx, &m.cfg.machineCfg)
	if err != nil {
		m.logger.Errorf("PutMachineConfiguration returned %s", resp.Error())
		return err
	}

	m.logger.Debug("PutMachineConfiguration returned")
	err = m.refreshMachineConfig()
	if err != nil {
		log.Errorf("Unable to inspect Firecracker MachineConfig. Continuing anyway. %s", err)
	}
	m.logger.Debug("createMachine returning")
	return err
}

func (m *Machine) createBootSource(ctx context.Context, imagePath, kernelArgs string) error {
	bsrc := models.BootSource{
		KernelImagePath: &imagePath,
		BootArgs:        kernelArgs,
	}

	resp, err := m.client.PutGuestBootSource(ctx, &bsrc)
	if err == nil {
		m.logger.Printf("PutGuestBootSource: %s", resp.Error())
	}

	return err
}

func (m *Machine) createNetworkInterface(ctx context.Context, iface NetworkInterface, iid int) error {
	ifaceID := strconv.Itoa(iid)
	m.logger.Printf("Attaching NIC %s (hwaddr %s) at index %s", iface.HostDevName, iface.MacAddress, ifaceID)

	ifaceCfg := models.NetworkInterface{
		IfaceID:           &ifaceID,
		GuestMac:          iface.MacAddress,
		HostDevName:       iface.HostDevName,
		State:             models.DeviceStateAttached,
		AllowMmdsRequests: iface.AllowMDDS,
	}

	resp, err := m.client.PutGuestNetworkInterfaceByID(ctx, ifaceID, &ifaceCfg)
	if err == nil {
		m.logger.Printf("PutGuestNetworkInterfaceByID: %s", resp.Error())
	}

	return err
}

// attachDrive attaches a secondary block device
func (m *Machine) attachDrive(ctx context.Context, dev BlockDevice, index int, root bool) error {
	var err error

	_, err = os.Stat(dev.HostPath)
	if err != nil {
		return err
	}

	readOnly := true

	switch dev.Mode {
	case "ro":
		readOnly = true
	case "rw":
		readOnly = false
	default:
		return errors.New("invalid drive permissions")
	}

	driveID := strconv.Itoa(index)
	d := models.Drive{
		DriveID:      &driveID,
		PathOnHost:   &dev.HostPath,
		IsRootDevice: &root,
		IsReadOnly:   &readOnly,
	}

	if len(m.cfg.RootPartitionUUID) > 0 && root {
		d.Partuuid = m.cfg.RootPartitionUUID
	}

	log.Infof("Attaching drive %s, mode %s, slot %s, root %t.", dev.HostPath, dev.Mode, driveID, root)

	respNoContent, err := m.client.PutGuestDriveByID(ctx, driveID, &d)
	if err == nil {
		m.logger.Printf("Attached drive %s: %s", dev.HostPath, respNoContent.Error())
	} else {
		m.logger.Errorf("Attach drive failed: %s: %s", dev.HostPath, err)
	}
	return err
}

func (m *Machine) attachRootDrive(ctx context.Context, dev BlockDevice) error {
	return m.attachDrive(ctx, dev, 1, true)
}

// addVsock adds a vsock to the instance
func (m *Machine) addVsock(ctx context.Context, dev VsockDevice) error {
	vsockCfg := models.Vsock{
		GuestCid: int64(dev.CID),
		ID:       &dev.Path,
	}
	resp, _, err := m.client.PutGuestVsockByID(ctx, dev.Path, &vsockCfg)
	if err != nil {
		return err
	}
	m.logger.Debugf("Attach vsock %s successful: %s", dev.Path, resp.Error())
	return nil
}

// StartInstance starts the firecracker VM
func (m *Machine) StartInstance(ctx context.Context) error {
	return m.startInstance(ctx)
}

func (m *Machine) startInstance(ctx context.Context) error {
	info := models.InstanceActionInfo{
		ActionType: models.InstanceActionInfoActionTypeInstanceStart,
	}

	resp, err := m.client.CreateSyncAction(ctx, &info)
	if err == nil {
		m.logger.Printf("startInstance successful: %s", resp.Error())
	} else {
		m.logger.Errorf("Starting instance: %s", err)
	}
	return err
}

// SetMetadata sets the machine's metadata for MDDS
func (m *Machine) SetMetadata(ctx context.Context, metadata interface{}) error {
	respcreated, respnocontent, err := m.client.PutMmds(ctx, metadata)

	if err == nil {
		var message string
		if respcreated != nil {
			message = respcreated.Error()
		}
		if respnocontent != nil {
			message = respnocontent.Error()
		}
		m.logger.Printf("SetMetadata successful: %s", message)
	} else {
		m.logger.Errorf("Setting metadata: %s", err)
	}
	return err
}

// refreshMachineConfig synchronizes our cached representation of the machine configuration
// with that reported by the Firecracker API
func (m *Machine) refreshMachineConfig() error {
	resp, err := m.client.GetMachineConfig()
	if err != nil {
		return err
	}

	m.logger.Infof("refreshMachineConfig: %s", resp.Error())
	m.machineConfig = *resp.Payload
	return nil
}

// waitForSocket waits for the given file to exist
func (m *Machine) waitForSocket(timeout time.Duration, exitchan chan error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error)
	ticker := time.NewTicker(10 * time.Millisecond)

	go func() {
		for {
			select {
			case <-ctx.Done():
				done <- ctx.Err()
				return
			case err := <-exitchan:
				done <- err
				return
			case <-ticker.C:
				if _, err := os.Stat(m.cfg.SocketPath); err != nil {
					continue
				}

				// Send test HTTP request to make sure socket is available
				if _, err := m.client.GetMachineConfig(); err != nil {
					continue
				}

				done <- nil
				return
			}
		}
	}()

	return <-done
}
