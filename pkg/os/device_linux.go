package os

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/emc-advanced-dev/pkg/errors"

	log "github.com/sirupsen/logrus"
)

func Mount(device BlockDevice) (mntpoint string, err error) {
	return MountDevice(device.Name())
}
func MountDevice(device string) (mntpoint string, err error) {
	defer func() {
		if err != nil {
			os.Remove(mntpoint)
		}
	}()

	mntpoint, err = ioutil.TempDir("", "stgr.mntpoint.")
	if err != nil {
		return
	}
	err = RunLogCommand("mount", device, mntpoint)
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

func (m *MsDosPartioner) Makebootable(partnum int) error {
	_, err := runParted(m.Device, "set", fmt.Sprintf("%d", partnum), "boot", "on")
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
			part = NewPartLoDevice(device.Name(), sectorsStart, sectorsSize)
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

func randomDeviceName() string {
	return "dev" + RandStringBytes(4)
}

// TODO: change this to api; like in here: https://www.versioneye.com/python/losetup/2.0.7 or here https://github.com/karelzak/util-linux/blob/master/sys-utils/losetup.c
type LoDevice struct {
	device        string
	createdDevice BlockDevice
	offset        DiskSize
	size          DiskSize
}

func NewLoDevice(device string) Resource {
	return &LoDevice{device, BlockDevice(""), nil, nil}
}
func NewPartLoDevice(device string, offset DiskSize, size DiskSize) Part {
	return &LoDevice{device, BlockDevice(""), offset, size}
}

func (p *LoDevice) Acquire() (BlockDevice, error) {
	log.WithFields(log.Fields{"cmd": "losetup", "device": p.device}).Debug("running losetup -f")

	args := []string{"-f", "--show", p.device}

	if p.size != nil {
		args = append(args, "--sizelimit", fmt.Sprintf("%d", p.size.ToBytes()))
	}

	if p.offset != nil {
		args = append(args, "--offset", fmt.Sprintf("%d", p.offset.ToBytes()))
	}

	out, err := exec.Command("losetup", args...).CombinedOutput()

	if err != nil {
		log.WithFields(log.Fields{"cmd": "losetup", "out": string(out), "device": p.device}).Debug("losetup -f failed")
		return BlockDevice(""), err
	}
	outString := strings.TrimSpace(string(out))
	p.createdDevice = BlockDevice(outString)
	return p.createdDevice, nil
}

func (p *LoDevice) Get() BlockDevice {
	return p.createdDevice
}

func (p *LoDevice) Release() error {
	return RunLogCommand("losetup", "-d", p.createdDevice.Name())
}

func (p *LoDevice) Size() DiskSize {
	return p.size
}

func (p *LoDevice) Offset() DiskSize {
	return p.offset
}
