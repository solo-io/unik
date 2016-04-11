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
            if (args.length != 7) {
                System.err.println("Usage: java VmAttachDisk <url> " +
                        "<username> <password> <vmname> <vmdkPath> <deviceKey>");
                System.exit(-1);
            }

            String vmname = args[4];
            String vmdkPath = args[5];
            int deviceKey = Integer.parseInt(args[6]);

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

            int scsiKey = -1;
            for (VirtualDevice vd : vm.getConfig().getHardware().getDevice()) {
                if (vd instanceof VirtualSCSIController) {
                    VirtualSCSIController vscsi = (VirtualSCSIController) vd;
                    System.out.println("found scsi controller:"+vscsi.getScsiCtlrUnitNumber()+" "+vscsi.getUnitNumber()+" "+vscsi.getKey());
                    scsiKey = vscsi.getKey();
                }
            }
            if (scsiKey == -1) {
                System.out.println("could not find scsi controller device on ");
                System.exit(-1);
            }

            VirtualMachineConfigSpec vmConfigSpec = new VirtualMachineConfigSpec();

            // mode: persistent|independent_persistent,independent_nonpersistent
            String diskMode = "persistent";
            VirtualDeviceConfigSpec vdiskSpec = createExistingDiskSpec(vmdkPath, scsiKey, deviceKey, diskMode);
            VirtualDeviceConfigSpec[] vdiskSpecArray = {vdiskSpec};
            vmConfigSpec.setDeviceChange(vdiskSpecArray);

            Task task = vm.reconfigVM_Task(vmConfigSpec);
            System.out.println(task.waitForTask());
            if (task.getTaskInfo() != null && task.getTaskInfo().getDescription() != null) {
                System.out.println(task.getTaskInfo().getDescription().getMessage());
            }
        }
        if (args[0].equals("CopyFile")) {
            if (args.length != 6) {
                System.err.println("Usage: java CopyFile <url> " +
                        "<username> <password> <sourcePath> <destinationPath>");
                System.exit(-1);
            }

            ServiceInstance si = new ServiceInstance(
                    new URL(args[1]), args[2], args[3], true);

            Datacenter datacenter = (Datacenter) new InventoryNavigator(si.getRootFolder()).searchManagedEntity("Datacenter", "ha-datacenter");

            String sourcePath = args[4];
            String destinationPath = args[5];

            FileManager fileManager = si.getFileManager();
            if (fileManager == null) {
                System.err.println("filemanager not available");
                System.exit(-1);
            }
            Task copyTask = fileManager.copyDatastoreFile_Task(sourcePath, datacenter, destinationPath, datacenter, true);

            System.out.println(copyTask.waitForTask());
            if (copyTask.getTaskInfo() != null && copyTask.getTaskInfo().getDescription() != null) {
                System.out.println(copyTask.getTaskInfo().getDescription().getMessage());
            }
        }
        if (args[0].equals("CopyVirtualDisk")) {
            if (args.length != 6) {
                System.out.println("Usage: java CopyVirtualDisk "
                        + "<url> <username> <password> <src> <dest>");
                System.exit(-1);
            }

            ServiceInstance si = new ServiceInstance(
                    new URL(args[1]), args[2], args[3], true);

            Datacenter dc = (Datacenter) new InventoryNavigator(
                    si.getRootFolder()).searchManagedEntity(
                    "Datacenter", "ha-datacenter");

            VirtualDiskManager diskManager = si.getVirtualDiskManager();
            if (diskManager == null) {
                System.out.println("DiskManager not available.");
                si.getServerConnection().logout();
                System.exit(-1);
            }

            String srcPath = args[4];
            String dstPath = args[5];
            VirtualDiskSpec copyDiskSpec = new VirtualDiskSpec();
            copyDiskSpec.setDiskType(VirtualDiskType.thin.name());
            copyDiskSpec.setAdapterType(VirtualDiskAdapterType.ide.name());
            Task cTask = diskManager.copyVirtualDisk_Task(srcPath, dc, dstPath, dc, copyDiskSpec, new Boolean(true));

            if (cTask.waitForTask().equals(Task.SUCCESS)) {
                System.out.println("Disk copied successfully!");
            } else {
                System.out.println("Disk copy failed!");
                return;
            }
            si.getServerConnection().logout();
        }
    }

    static VirtualDeviceConfigSpec createExistingDiskSpec(String fileName, int controllerKey, int deviceKey, String diskMode) {
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
}

