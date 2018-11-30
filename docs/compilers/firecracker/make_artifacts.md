# How to create the root fs this copiler uses

## Create an image file
```
dd if=/dev/zero of=rootfs bs=1M count=50
mkfs.ext4 rootfs
```

## Mount it
```
mkdir tmp
sudo mount rootfs tmp -o loop
```

## Setup a base:
```
wget http://dl-cdn.alpinelinux.org/alpine/v3.8/releases/x86_64/alpine-minirootfs-3.8.1-x86_64.tar.gz
cd tmp
sudo tar xzf ../alpine-minirootfs-3.8.1-x86_64.tar.gz
cd ..
```

## Setup various settings
This will setup dns for netowrk, and an inittab that will start start-script.

```
echo nameserver 1.1.1.1 | sudo tee ./tmp/etc/resolv.conf
cat interfaces | sudo tee ./tmp/etc/network/interfaces
cat inittab | sudo tee ./tmp/etc/inittab
cat start-script | sudo tee ./tmp/start.sh
```


## Update the image from the inside

Install openrc, set root password (makes debugging easier).
```
sudo chroot tmp/ /bin/sh
passwd root -d root
apk update
apk add openrc
exit
```

## Cleanup
```
sudo umount tmp
rmdir tmp
```

# Get a kernel from linuxkit:
```
VERSION=4.19.4
CID=$(docker create linuxkit/kernel:$VERSION dummycommand)
docker cp $CID:/kernel ./kernel-$VERSION
docker rm $CID
wget https://raw.githubusercontent.com/torvalds/linux/master/scripts/extract-vmlinux
bash extract-vmlinux kernel-$VERSION > kernel-$VERSION-elf
```