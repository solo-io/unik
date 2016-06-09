package com.emc.unik;

import java.net.URL;

import com.vmware.vim25.*;
import com.vmware.vim25.mo.*;

public class VmAttachDisk {
    public static void main(String[] args) throws Exception {
        if (args.length < 1) {
            System.err.println("Usage: java VmAttachDisk|CopyFile|CopyVirtualDisk [<opts>]");
            System.exit(-1);
        }

        if (args[0].equals("VmAttachDisk")) {
            if (args.length != 8) {
                System.err.println("Usage: java VmAttachDisk <url> " +
                        "<username> <password> <vmname> <vmdkPath> <DeviceType: SCSI|IDE> <deviceSlot>");
                System.exit(-1);
            }

            String vmname = args[4];
            String vmdkPath = args[5];
            String deviceType = args[6];
            int deviceSlot = Integer.parseInt(args[7]);

            ServiceInstance si = new ServiceInstance(
                    new URL(args[1]), args[2], args[3], true);

            Folder rootFolder = si.getRootFolder();
            VirtualMachine vm = (VirtualMachine) new InventoryNavigator(
                    rootFolder).searchManagedEntity("VirtualMachine", vmname);

            if (vm == null) {
                System.out.println("No VM " + vmname + " found");
                si.getServerConnection().logout();
                System.exit(-1);
            }

            int storageDeviceKey = -1;
            if (deviceType.contains("SCSI")) {
                storageDeviceKey = getScsiDeviceKey(vm);
            } else if (deviceType.contains("IDE")) {
                storageDeviceKey = getIdeDeviceKey(vm);
            }

            if (storageDeviceKey == -1) {
                System.out.println("could not find controller device type: "+deviceType);
                System.exit(-1);
            }

            VirtualMachineConfigSpec vmConfigSpec = new VirtualMachineConfigSpec();

            // mode: persistent|independent_persistent,independent_nonpersistent
            String diskMode = "persistent";
            VirtualDeviceConfigSpec vdiskSpec = createExistingDiskSpec(vmdkPath, storageDeviceKey, deviceSlot, diskMode);
            VirtualDeviceConfigSpec[] vdiskSpecArray = {vdiskSpec};
            vmConfigSpec.setDeviceChange(vdiskSpecArray);

            Task task = vm.reconfigVM_Task(vmConfigSpec);
            System.out.println(task.waitForTask());
            if (task.getTaskInfo() != null && task.getTaskInfo().getDescription() != null) {
                System.out.println(task.getTaskInfo().getDescription().getMessage());
                if (task.getTaskInfo().getDescription().getMessage().contains("success")) {
                    return;
                }
                System.exit(-1);
            }
        }

        if (args[0].equals("VmDetachDisk")) {
            if (args.length != 7) {
                System.err.println("Usage: java VmAttachDisk <url> " +
                        "<username> <password> <vmname>  <DeviceType: SCSI|IDE> <deviceSlot>");
                System.exit(-1);
            }

            String vmname = args[4];
            String deviceType = args[6];
            int deviceSlot = Integer.parseInt(args[6]);

            ServiceInstance si = new ServiceInstance(
                    new URL(args[1]), args[2], args[3], true);

            Folder rootFolder = si.getRootFolder();
            VirtualMachine vm = (VirtualMachine) new InventoryNavigator(
                    rootFolder).searchManagedEntity("VirtualMachine", vmname);

            if (vm == null) {
                System.out.println("No VM " + vmname + " found");
                si.getServerConnection().logout();
                System.exit(-1);
            }

            int storageDeviceKey = -1;
            if (deviceType.contains("SCSI")) {
                storageDeviceKey = getScsiDeviceKey(vm);
            } else if (deviceType.contains("IDE")) {
                storageDeviceKey = getIdeDeviceKey(vm);
            }

            if (storageDeviceKey == -1) {
                System.out.println("could not find controller device type: "+deviceType);
                System.exit(-1);
            }

            VirtualMachineConfigSpec vmConfigSpec = new VirtualMachineConfigSpec();

            VirtualDeviceConfigSpec vdiskSpec = createRemoveDiskSpec(storageDeviceKey, deviceSlot);
            VirtualDeviceConfigSpec[] vdiskSpecArray = {vdiskSpec};
            vmConfigSpec.setDeviceChange(vdiskSpecArray);

            Task task = vm.reconfigVM_Task(vmConfigSpec);
            System.out.println(task.waitForTask());
            if (task.getTaskInfo() != null && task.getTaskInfo().getDescription() != null) {
                System.out.println(task.getTaskInfo().getDescription().getMessage());
            }
            if (task.getTaskInfo().getDescription().getMessage().contains("success")) {
                return;
            }
            System.exit(-1);
        }
        if (args[0].equals("CopyFile")) {
            if (args.length != 7) {
                System.err.println("Usage: java CopyFile <url> " +
                        "<username> <password> <datacenter> <sourcePath> <destinationPath>");
                System.exit(-1);
            }

            ServiceInstance si = new ServiceInstance(
                    new URL(args[1]), args[2], args[3], true);
            String datacenterName = args[4];

            Datacenter datacenter = (Datacenter) new InventoryNavigator(si.getRootFolder()).searchManagedEntity("Datacenter", datacenterName);

            String sourcePath = args[5];
            String destinationPath = args[6];

            FileManager fileManager = si.getFileManager();
            if (fileManager == null) {
                System.err.println("filemanager not available");
                System.exit(-1);
            }
            Task copyTask = fileManager.copyDatastoreFile_Task(sourcePath, datacenter, destinationPath, datacenter, true);

            String res = copyTask.waitForTask();
            System.out.println(res);
            if (res.contains("success")) {
                return;
            }
            System.exit(-1);
        }
        if (args[0].equals("CopyVirtualDisk")) {
            if (args.length != 7) {
                System.out.println("Usage: java CopyVirtualDisk "
                        + "<url> <username> <password> <datacenter> <src> <dest>");
                System.exit(-1);
            }

            ServiceInstance si = new ServiceInstance(
                    new URL(args[1]), args[2], args[3], true);

            String datacenterName = args[4];

            Datacenter dc = (Datacenter) new InventoryNavigator(
                    si.getRootFolder()).searchManagedEntity(
                    "Datacenter", datacenterName);

            VirtualDiskManager diskManager = si.getVirtualDiskManager();
            if (diskManager == null) {
                System.out.println("DiskManager not available.");
                si.getServerConnection().logout();
                System.exit(-1);
            }

            String srcPath = args[5];
            String dstPath = args[6];
            VirtualDiskSpec copyDiskSpec = new VirtualDiskSpec();
            copyDiskSpec.setDiskType(VirtualDiskType.thin.name());
            copyDiskSpec.setAdapterType(VirtualDiskAdapterType.ide.name());
            Task cTask = null;
            try {
                cTask = diskManager.copyVirtualDisk_Task(srcPath, dc, dstPath, dc, copyDiskSpec, new Boolean(true));
            } catch (InvalidArgument e){
                e.getInvalidProperty();
                e.printStackTrace();
                System.exit(-1);
            }

            if (cTask.waitForTask().equals(Task.SUCCESS)) {
                System.out.println("Disk copied successfully!");
            } else {
                System.out.println("Disk copy failed!");
                System.out.println(cTask.getTaskInfo().getError().getLocalizedMessage());
                System.exit(-1);
            }
            si.getServerConnection().logout();
        }
    }

    private static VirtualDeviceConfigSpec createExistingDiskSpec(String fileName, int controllerKey, int deviceKey, String diskMode) {
        VirtualDeviceConfigSpec diskSpec =
                new VirtualDeviceConfigSpec();
        diskSpec.setOperation(VirtualDeviceConfigSpecOperation.add);
        // do not set diskSpec.fileOperation!
        VirtualDisk vd = new VirtualDisk();
        vd.setCapacityInKB(-1);
        vd.setKey(deviceKey);
        vd.setUnitNumber(new Integer(deviceKey));
        vd.setControllerKey(new Integer(controllerKey));
        VirtualDiskFlatVer2BackingInfo diskfileBacking =
                new VirtualDiskFlatVer2BackingInfo();
        diskfileBacking.setFileName(fileName);
        diskfileBacking.setDiskMode(diskMode);
        vd.setBacking(diskfileBacking);
        diskSpec.setDevice(vd);
        return diskSpec;
    }

    private static int getScsiDeviceKey(VirtualMachine vm) {
        for (VirtualDevice vd : vm.getConfig().getHardware().getDevice()) {
            if (vd instanceof VirtualSCSIController) {
                VirtualSCSIController vscsi = (VirtualSCSIController) vd;
                System.out.println("found scsi controller:"+vscsi.getScsiCtlrUnitNumber()+" "+vscsi.getUnitNumber()+" "+vscsi.getKey());
                return vscsi.getKey();
            }
        }
        return -1;
    }

    private static int getIdeDeviceKey(VirtualMachine vm) {
        for (VirtualDevice vd : vm.getConfig().getHardware().getDevice()) {
            if (vd instanceof VirtualIDEController) {
                VirtualIDEController vide = (VirtualIDEController) vd;
                System.out.println("found ide controller:"+" "+vide.getUnitNumber()+" "+vide.getKey());
                return vide.getKey();
            }
        }
        return -1;
    }

    private static VirtualDeviceConfigSpec createRemoveDiskSpec(int controllerKey, int deviceKey) {
        VirtualDeviceConfigSpec diskSpec =
                new VirtualDeviceConfigSpec();
        diskSpec.setOperation(VirtualDeviceConfigSpecOperation.remove);

        VirtualDisk vd = new VirtualDisk();
        vd.setCapacityInKB(-1);
        vd.setKey(deviceKey);
        vd.setUnitNumber(new Integer(deviceKey));
        vd.setControllerKey(new Integer(controllerKey));
        diskSpec.setDevice(vd);
        return diskSpec;
    }
}

