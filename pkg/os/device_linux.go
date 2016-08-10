package os

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/emc-advanced-dev/pkg/errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func Mount(device BlockDevice) (mntpoint string, err error) {
	defer func() {
		if err != nil {
			os.Remove(mntpoint)
		}
	}()

	mntpoint, err = ioutil.TempDir("", "stgr.mntpoint.")
	if err != nil {
		return
	}
	err = RunLogCommand("mount", device.Name(), mntpoint)
	return
}

func Umount(point string) error {

	err := RunLogCommand("umount", point)
	if err != nil {
		return err
	}
	// ignore errors.
	err = os.Remove(point)
	if err != nil {
		log.WithField("err", err).Warn("umount rmdir failed")
	}

	return nil
}

func runParted(device string, args ...string) ([]byte, error) {
	log.WithFields(log.Fields{"device": device, "args": args}).Debug("running parted")
	args = append([]string{"--script", "--machine", device}, args...)
	out, err := exec.Command("parted", args...).CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"args": args, "err": err, "out": string(out)}).Error("parted failed")
	}
	return out, err
}

type MsDosPartioner struct {
	Device string
}

func (m *MsDosPartioner) MakeTable() error {
	_, err := runParted(m.Device, "mklabel", "msdos")
	return err
}

func (m *MsDosPartioner) MakePart(partType string, start, size DiskSize) error {
	_, err := runParted(m.Device, "mkpart", partType, start.ToPartedFormat(), size.ToPartedFormat())
	return err
}
func (m *MsDosPartioner) MakePartTillEnd(partType string, start DiskSize) error {
	_, err := runParted(m.Device, "mkpart", partType, start.ToPartedFormat(), "100%")
	return err
}

type DiskLabelPartioner struct {
	Device string
}

func (m *DiskLabelPartioner) MakeTable() error {
	_, err := runParted(m.Device, "mklabel", "bsd")
	return err
}

func (m *DiskLabelPartioner) MakePart(partType string, start, size DiskSize) error {
	_, err := runParted(m.Device, "mkpart", partType, start.ToPartedFormat(), size.ToPartedFormat())
	return err
}

func ListParts(device BlockDevice) ([]Part, error) {
	var parts []Part
	out, err := runParted(device.Name(), "unit B", "print")
	if err != nil {
		return parts, nil
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	/* example output

	  BYT;
	  /dev/xvda:42949672960B:xvd:512:512:msdos:Xen Virtual Block Device;
	  1:8225280B:42944186879B:42935961600B:ext4::boot;

	  ================

	  BYT;
	  /home/ubuntu/yuval:1073741824B:file:512:512:bsd:;
	  1:2097152B:99614719B:97517568B:::;
	  2:99614720B:200278015B:100663296B:::;
	  3:200278016B:299892735B:99614720B:::;

	========= basically:
	device:size:
	partnum:start:end:size

	*/

	// skip to the parts..
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, device.Name()) {
			break
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, ":")

		partNum, err := strconv.ParseInt(tokens[0], 0, 0)
		if err != nil {
			return parts, err
		}

		start, err := getByteNumber(tokens[1])
		if err != nil {
			return parts, err
		}

		end, err := getByteNumber(tokens[2])
		if err != nil {
			return parts, err
		}

		size, err := getByteNumber(tokens[3])
		if err != nil {
			return parts, err
		}

		//validate Part is consistent:
		if end-start != size-1 {
			log.WithFields(log.Fields{"start": start, "end": end, "size": size}).Error("Sizes not consistent")
			return parts, errors.New("Sizes are inconsistent. part not continous?", nil)
		}

		var part Part
		partName := getDevicePart(device.Name(), partNum)
		// part devices may or may not created the partition mappings. so deal with both options
		if _, err := os.Stat(partName); os.IsNotExist(err) {
			// device does not exist
			sectorsStart, err := ToSectors(start)
			if err != nil {
				return parts, err
			}
			sectorsSize, err := ToSectors(size)
			if err != nil {
				return parts, err
			}
			part = NewDMPartedPart(sectorsStart, sectorsSize, device, partNum)
		} else {
			// device exists
			var release func(BlockDevice) error = nil

			// prated might have created the mapping for us. unfortunatly it does not remove it...
			if strings.HasPrefix(partName, "/dev/mapper") {
				release = func(d BlockDevice) error {
					return RunLogCommand("dmsetup", "remove", d.Name())
				}
			}

			part = &PartedPart{BlockDevice(partName), start, size, release}
		}
		parts = append(parts, part)
	}

	return parts, nil
}

func getDevicePart(device string, part int64) string {
	return fmt.Sprintf("%s%c", device, '0'+part)
}

func getByteNumber(token string) (Bytes, error) {
	tokenLen := len(token)
	if tokenLen == 0 {
		return 0, errors.New("Not a number", nil)
	}
	// remove the B

	if token[tokenLen-1] != 'B' {

		return 0, errors.New("Unknown unit for number", nil)
	}

	res, err := strconv.ParseInt(token[:tokenLen-1], 0, 0)
	return Bytes(res), err
}

type PartedPart struct {
	Device  BlockDevice
	offset  DiskSize
	size    DiskSize
	release func(BlockDevice) error
}

func (p *PartedPart) Size() DiskSize {
	return p.size
}
func (p *PartedPart) Offset() DiskSize {
	return p.offset
}

func (p *PartedPart) Acquire() (BlockDevice, error) {

	return p.Get(), nil
}

func (p *PartedPart) Release() error {
	if p.release != nil {
		return p.release(p.Device)
	}
	return nil
}

func (p *PartedPart) Get() BlockDevice {
	return p.Device
}

type DeviceMapperDevice struct {
	DeviceName string

	start Sectors
	size  Sectors

	orginalDevice BlockDevice
}

func randomDeviceName() string {
	return "dev" + RandStringBytes(4)
}

// Device device name is generated, user can chagne it..
func NewDMPartedPart(start, size Sectors, device BlockDevice, partNum int64) Part {
	name := randomDeviceName()
	newDeviceName := fmt.Sprintf("%s%c", name, '0'+partNum)
	return &DeviceMapperDevice{newDeviceName, start, size, device}
}

func NewDevice(start, size Sectors, origDevice BlockDevice, deivceName string) Resource {
	return &DeviceMapperDevice{deivceName, start, size, origDevice}
}

func (p *DeviceMapperDevice) Size() DiskSize {
	return p.size
}
func (p *DeviceMapperDevice) Offset() DiskSize {
	return p.start
}

func (p *DeviceMapperDevice) Acquire() (BlockDevice, error) {
	// dmsetup create partition${PARTI} --table "0 $SIZE linear $DEVICE $SECTOR"
	table := fmt.Sprintf("0 %d linear %s %d", p.size, p.orginalDevice, p.start)

	err := RunLogCommand("dmsetup", "create", p.DeviceName, "--table", table)

	if err == nil && !IsExists(p.Get().Name()) {
		err = RunLogCommand("dmsetup", "mknodes", p.DeviceName)
	}

	return p.Get(), err
}

func (p *DeviceMapperDevice) Release() error {
	err := RunLogCommand("dmsetup", "remove", p.DeviceName)
	if err == nil && IsExists(p.Get().Name()) {
		err = os.Remove(p.Get().Name())
	}
	return err

}

func (p *DeviceMapperDevice) Get() BlockDevice {
	newDevice := "/dev/mapper/" + p.DeviceName
	return BlockDevice(newDevice)
}

// TODO: change this to api; like in here: https://www.versioneye.com/python/losetup/2.0.7 or here https://github.com/karelzak/util-linux/blob/master/sys-utils/losetup.c
type LoDevice struct {
	device        string
	createdDevice BlockDevice
}

func NewLoDevice(device string) Resource {
	return &LoDevice{device, BlockDevice("")}
}

func (p *LoDevice) Acquire() (BlockDevice, error) {
	// dmsetup create partition${PARTI} --table "0 $SIZE linear $DEVICE $SECTOR"
	log.WithFields(log.Fields{"cmd": "losetup", "device": p.device}).Debug("running losetup -f")

	out, err := exec.Command("losetup", "-f", "--show", p.device).CombinedOutput()

	if err != nil {
		log.WithFields(log.Fields{"cmd": "losetup", "out": string(out), "device": p.device}).Debug("losetup -f failed")
		return BlockDevice(""), err
	}
	outString := strings.TrimSpace(string(out))
	p.createdDevice = BlockDevice(outString)
	return p.createdDevice, nil
}

func (p *LoDevice) Release() error {
	return RunLogCommand("losetup", "-d", p.createdDevice.Name())
}
