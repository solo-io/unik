# patch /opt/rumprun/lib/librumprun_base/config.c < /tmp/patch

set -e

cd  /opt/rumprun/

DESTDIR=/usr/local
BUILDRUMP_EXTRA=

cp /tmp/patches/buildrump.sh/brlib/libnetconfig/dhcp_configure.c /opt/rumprun/buildrump.sh/brlib/libnetconfig/dhcp_configure.c

./build-rr.sh -d $DESTDIR -o ./obj $PLATFORM build -- $BUILDRUMP_EXTRA && \
./build-rr.sh -d $DESTDIR -o ./obj $PLATFORM install