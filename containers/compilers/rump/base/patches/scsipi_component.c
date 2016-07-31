/*	$NetBSD: scsipi_component.c,v 1.2 2016/01/26 23:12:16 pooka Exp $	*/

#include <sys/param.h>
#include <sys/conf.h>
#include <sys/device.h>
#include <sys/kmem.h>
#include <sys/stat.h>
#include <sys/disklabel.h>

#include "ioconf.c"

#include <rump-sys/kern.h>
#include <rump-sys/vfs.h>

RUMP_COMPONENT(RUMP_COMPONENT_DEV)
{
	extern struct bdevsw sd_bdevsw, cd_bdevsw;
	extern struct cdevsw sd_cdevsw, cd_cdevsw;
    int i;
	devmajor_t bmaj, cmaj;
    
    char bDevice[] = "/dev/sd0";
    char cDevice[] = "/dev/rsd0";

	config_init_component(cfdriver_ioconf_scsipi,
	    cfattach_ioconf_scsipi, cfdata_ioconf_scsipi);

	bmaj = cmaj = -1;
	FLAWLESSCALL(devsw_attach("sd", &sd_bdevsw, &bmaj, &sd_cdevsw, &cmaj));

    for (i=0; i < 8; i++) {
        bDevice[7] = '0' + i;
        cDevice[8] = '0' + i;
    
        FLAWLESSCALL(rump_vfs_makedevnodes(S_IFBLK, bDevice, 'a',
            bmaj, MAXPARTITIONS*i, 8));
        FLAWLESSCALL(rump_vfs_makedevnodes(S_IFCHR, cDevice, 'a',
            cmaj, MAXPARTITIONS*i, 8));
    }
    
	bmaj = cmaj = -1;
	FLAWLESSCALL(devsw_attach("cd", &cd_bdevsw, &bmaj, &cd_cdevsw, &cmaj));

	FLAWLESSCALL(rump_vfs_makedevnodes(S_IFBLK, "/dev/cd0", 'a',
	    bmaj, 0, 8));
	FLAWLESSCALL(rump_vfs_makedevnodes(S_IFCHR, "/dev/rcd0", 'a',
	    cmaj, 0, 8));
}
