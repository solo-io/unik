Debugging UniK Amazon images

When debugging unikernels, the main tool amazon offers us is "Get System logs". This basically limits you to "printf" debugging style.

To use a real debugger, you can build your own xen hypervisor, that will run the same unik images that were built for amazon. 
You can then use gdb to debug your code.

These instructions are for OS X. can be used in Linux with minor modifications.

# Get Xen

I used this vagrant box: https://github.com/englishm/vagrant-xen

In the vagrant config, forward port 9999:
```
...
config.vm.network "forwarded_port", guest: 9999, host: 9999
...

```

This will be important later when we want to connect to this machine with gdb.

## PV Grub
Once you got the box running, ssh inside it (```vagrant ssh```). 
AWS uses pv grub too boot images, you will need to build pv grub (as it's not there by default).

In general, the instructions are [here](http://wiki.xen.org/wiki/PvGrub
). Before doing "./configure", install these packages as well (otherwise build will fail):

```
sudo apt-get install libaio-dev libssl-dev libc6-dev-i386 texinfo git
```

## Add a bridge

To add a bridge, add the following lines to /etc/network/interfaces:
```
iface xenbr0 inet dhcp
    bridge_ports eth0
    bridge_stp off
    bridge_maxwait 0
    bridge_fd 0
```

Then run this (do this again if you restart, not sure why, but it is not automatic):

```
sudo ifdown eth0 && sudo ifup xenbr0 && sudo ifup eth0
```

Sources:
- http://askubuntu.com/questions/136089/how-to-set-up-bridged-networking-in-xen
- https://help.ubuntu.com/community/Xen



# Fake AWS metadata service

Create a xen script on this file "/etc/xen/scripts/metadata-fake", with the following. 
Change 10.0.2.15 to your machine's IP

```
#!/bin/bash

dir=$(dirname "$0")
. "$dir/vif-bridge"
case "$command" in
    add|online) 
            # TODO support -i $dev so this can be used for multiple vms; it's not working from some reason
            iptables -t nat -A PREROUTING   -d 169.254.169.254 -j DNAT --to-destination 10.0.2.15 
        ;;
    remove|offline)
            iptables -t nat -D PREROUTING   -d 169.254.169.254 -j DNAT --to-destination 10.0.2.15 || : 
        ;;
esac
```

And of course

```chmod a+x /etc/xen/scripts/metadata-fake```

## Run Metadata Server

Unik expects a string-to-string map of environment variables in the user-data.
We'll just create an empty map:

```
mkdir  latest
cat > latest/user-data <<EOF
{}
EOF
```

Then start python fake metadata server:
```
sudo python -m SimpleHTTPServer 80
```

# XL Config file


```
# Example PV Linux guest configuration
# =====================================================================
#
# This is a fairly minimal example of what is required for a
# Paravirtualised Linux guest. For a more complete guide see xl.cfg(5)

# Guest name
name = "aws-test"

# 128-bit UUID for the domain as a hexadecimal number.
# Use "uuidgen" to generate one if required.
# The default behavior is to generate a new UUID each time the guest is started.
#uuid = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"

kernel = "/home/vagrant/xen/dist/install/usr/local/lib/xen/boot/pv-grub-x86_64.gz"
extra = "(hd0)/boot/grub/menu.lst"

# Initial memory allocation (MB)
memory = 1024

# Maximum memory (MB)
# If this is greater than `memory' then the slack will start ballooned
# (this assumes guest kernel support for ballooning)
#maxmem = 512

# Number of VCPUS
vcpus = 1

# Network devices
# A list of 'vifspec' entries as described in
# docs/misc/xl-network-configuration.markdown
vif = [ 'bridge=xenbr0,script=metadata-fake,mac=00:16:3e:58:88:57' ]

# Disk Devices
# A list of `diskspec' entries as described in
# docs/misc/xl-disk-configuration.txt
disk = [ '/home/vagrant/boot-vol.img,raw,sda1,rw' ]
```

Save this as aws-test.conf

Notes:
- memory and vcpus should match the instance you are emulating
- disk should point the image built by unik. use "--no-cleanup" in `unik build` so it would not delete it after it's uploaded to AWS. 
You will see new file and a folder created in the unik tmp folder. For example:
```
~/W/g/s/g/e/unik ❯❯❯ ls -tlr  ~/.unik/tmp/
total 319488
...
drwx------  7 kohavy  720748206       238 Aug  3 16:47 bootable-image-directory.411462683
-rw-r--r--  1 kohavy  720748206  54525952 Aug  3 16:47 boot-creator-result.img.940436670
```
Copy the file (in our example, `boot-creator-result.img.940436670`) to the vagrant machine. note the folder as you will need it for later.
- kernel is the path to pv-grub built previously.
- leave extra param as it is (it describes how unik layouts the disk image)

# Run! 

```
sudo  xl create -c ./aws-test.conf
```

`ctrl+]` to exit console

You can delete the vm when you are done:
```
sudo xl destroy aws-test
```

# Debug! 
After starting your vm (with `xl create`), to attached with a debugger,  get dom id:

```
sudo  xl list
Name                                        ID   Mem VCPUs	State	Time(s)
Domain-0                                     0   837     1     r-----      12.3
aws-test                                     3  1024     1     --p---       0.0
```

Here the ID is 3. replace 3 with your dom id.

Start gdb stub on the vagrant machine
```
sudo /usr/lib/xen-4.4/bin/gdbsx -a 3 64 9999
```
OR just do this:
```
sudo /usr/lib/xen-4.4/bin/gdbsx -a $(sudo xl list|tail -1 | awk '{print $2}') 64 9999
```


Start our gdb container (your container tab might differ, check containers/versions.json):
```
docker run --net host --rm -t -i -v /Users/kohavy/.unik/tmp/bootable-image-directory.411462683/:/opt/code:ro  projectunik/rump-debugger-xen:7fa273029766
/opt/gdb-7.11/gdb/gdb -ex 'target remote 192.168.1.109:9999' /opt/code/program.bin
```

Debug your problems away!

Note: Bootable-image-directory.411462683 is the directory that the image was formed from. unik will keep it intact if you use "--no-cleanup".
This directory and the image in the XL config file *MUST* match for source level debugging to work!


If you connected with GDB in an early stage, grub might have not loaded ethe kernel yet.
I just place a breakpoint on ```_minios_hypercall_page``` and continued running a few times until the kernel was loaded.