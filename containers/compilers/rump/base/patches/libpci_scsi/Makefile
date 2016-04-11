#    $NetBSD: Makefile,v 1.5 2015/11/16 23:27:08 pooka Exp $
#

RUMPTOP=${TOPRUMP}

.PATH:    ${RUMPTOP}/../dev/pci ${RUMPTOP}/../dev/ ${RUMPTOP}/../dev/ic

LIB=    rumpdev_pci_scsi
COMMENT=PCI SCSI controller drivers

IOCONF=    PCI_SCSI.ioconf
RUMP_COMPONENT=ioconf

SRCS+= mpt_pci.c  mpt_netbsd.c mpt.c mpt_debug.c

CPPFLAGS+= -I${RUMPTOP}/librump/rumpkern -I${RUMPTOP}/../dev

.include "${RUMPTOP}/Makefile.rump"
.include <bsd.lib.mk>
.include <bsd.klinks.mk>
