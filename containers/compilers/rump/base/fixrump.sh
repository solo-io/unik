# patch /opt/rumprun/lib/librumprun_base/config.c < /tmp/patch

set -e

cd  /opt/rumprun/

DESTDIR=/usr/local
BUILDRUMP_EXTRA=

if [ "$PLATFORM" = "" ]; then
  echo PLATFORM should be xen or hw
  exit 1
fi


if [ "$PLATFORM" = "hw" ]; then
# ppb patch
cat >>  /opt/rumprun/src-netbsd/sys/rump/dev/lib/libpci/PCI.ioconf <<EOF
pci*    at ppb? bus ?
ppb*    at pci? dev ? function ?
EOF

# add ppb pride for vmware network cards
sed -i -e 's/SRCS+=	pci.c/SRCS+=	ppb.c pci.c/' /opt/rumprun/src-netbsd/sys/rump/dev/lib/libpci/Makefile

# add scsi driver for vmware hard drives
touch /opt/rumprun/src-netbsd/sys/dev/ic/bio.h

cp -r /tmp/patches/libpci_scsi        /opt/rumprun/src-netbsd/sys/rump/dev/lib/
cp    /tmp/patches/scsipi_component.c /opt/rumprun/src-netbsd/sys/rump/dev/lib/libscsipi/scsipi_component.c

sed -i -e 's/RUMPPCIDEVS+=\tmiiphy/RUMPPCIDEVS+=  pci_scsi miiphy/' /opt/rumprun/src-netbsd/sys/rump/dev/Makefile.rumpdevcomp

cp /tmp/patches/rumprun-bake.conf /opt/rumprun/app-tools/rumprun-bake.conf


cp /tmp/patches/rump/kern.ldscript   /opt/rumprun/platform/hw/arch/amd64/kern.ldscript
else
# platform is xen
cp /tmp/patches/rump/minios-x86_64.lds /opt/rumprun/platform/xen/xen/arch/x86/minios-x86_64.lds

fi


# all platforms

cp /tmp/patches/rump/rumprun.c   /opt/rumprun/lib/librumprun_base/rumprun.c
cp /tmp/patches/rump/rumprun-bake.in   /opt/rumprun/app-tools/rumprun-bake.in


./build-rr.sh -d $DESTDIR -o ./obj $PLATFORM build -- $BUILDRUMP_EXTRA && \
./build-rr.sh -d $DESTDIR -o ./obj $PLATFORM install
