/*-
 * Copyright (c) 2015 Antti Kantee.  All Rights Reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE AUTHOR ``AS IS'' AND ANY EXPRESS
 * OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE AUTHOR OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
 * LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
 * OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
 * SUCH DAMAGE.
 */

#include <sys/cdefs.h>

#include <sys/types.h>
#include <sys/mount.h>
#include <sys/queue.h>
#include <sys/sysctl.h>

#include <assert.h>
#include <err.h>
#include <errno.h>
#include <pthread.h>
#include <stdio.h>
#include <sched.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include <rump/rump.h>
#include <rump/rump_syscalls.h>

#include <fs/tmpfs/tmpfs_args.h>

#include <bmk-core/platform.h>

#include <rumprun-base/rumprun.h>
#include <rumprun-base/config.h>

#include "rumprun-private.h"

static pthread_mutex_t w_mtx;
static pthread_cond_t w_cv;

int rumprun_enosys(void);
int
rumprun_enosys(void)
{

	return ENOSYS;
}
__strong_alias(rumprun_notmain,rumprun_enosys);
__weak_alias(rumprun_main1,rumprun_notmain);
__weak_alias(rumprun_main2,rumprun_notmain);
__weak_alias(rumprun_main3,rumprun_notmain);
__weak_alias(rumprun_main4,rumprun_notmain);
__weak_alias(rumprun_main5,rumprun_notmain);
__weak_alias(rumprun_main6,rumprun_notmain);
__weak_alias(rumprun_main7,rumprun_notmain);
__weak_alias(rumprun_main8,rumprun_notmain);

__weak_alias(rump_init_server,rumprun_enosys);

int rumprun_cold = 1;

void
rumprun_boot(char *cmdline)
{
	struct tmpfs_args ta = {
		.ta_version = TMPFS_ARGS_VERSION,
		.ta_size_max = 1*1024*1024,
		.ta_root_mode = 01777,
	};
	int tmpfserrno;
	char *sysproxy;
	int rv, x;

	rump_boot_setsigmodel(RUMP_SIGMODEL_IGNORE);
	rump_init();

	/* mount /tmp before we let any userspace bits run */
	rump_sys_mount(MOUNT_TMPFS, "/tmp", 0, &ta, sizeof(ta));
	tmpfserrno = errno;

	/*
	 * XXX: _netbsd_userlevel_init() should technically be called
	 * in mainbouncer() per process.  However, there's currently no way
	 * to run it per process, and besides we need a fully functional
	 * libc to run sysproxy and rumprun_config(), so we just call it
	 * here for the time being.
	 *
	 * Eventually, we of course want bootstrap process which is
	 * rumprun() internally.
	 */
	rumprun_lwp_init();
	_netbsd_userlevel_init();

	/* print tmpfs result only after we bootstrapped userspace */
	if (tmpfserrno == 0) {
		fprintf(stderr, "mounted tmpfs on /tmp\n");
	} else {
		warnx("FAILED: mount tmpfs on /tmp: %s", strerror(tmpfserrno));
	}

	/*
	 * We set duplicate address detection off for
	 * immediately operational DHCP addresses.
	 * (note: we don't check for errors since net.inet.ip.dad_count
	 * is not present if the networking stack isn't present)
	 */
	x = 0;
	sysctlbyname("net.inet.ip.dad_count", NULL, NULL, &x, sizeof(x));

	rumprun_config(cmdline);

	sysproxy = getenv("RUMPRUN_SYSPROXY");
	if (sysproxy) {
		if ((rv = rump_init_server(sysproxy)) != 0)
			err(1, "failed to init sysproxy at %s", sysproxy);
		printf("sysproxy listening at: %s\n", sysproxy);
	}

	/*
	 * give all threads a chance to run, and ensure that the main
	 * thread has gone through a context switch
	 */
	sched_yield();

	pthread_mutex_init(&w_mtx, NULL);
	pthread_cond_init(&w_cv, NULL);

	rumprun_cold = 0;
}

/*
 * XXX: we have to use pthreads as the main threads for rumprunners
 * because otherwise libpthread goes haywire because it doesn't understand
 * the concept of multiple main threads (which is sort of understandable ...)
 */
#define RUMPRUNNER_DONE		0x10
#define RUMPRUNNER_DAEMON	0x20
struct rumprunner {
	int (*rr_mainfun)(int, char *[]);
	int rr_argc;
	char **rr_argv;

	pthread_t rr_mainthread;
	struct lwp *rr_lwp;

	int rr_flags;

	LIST_ENTRY(rumprunner) rr_entries;
};
static LIST_HEAD(,rumprunner) rumprunners = LIST_HEAD_INITIALIZER(&rumprunners);
static int rumprun_done;

/* XXX: does not yet nuke any pthread that mainfun creates */
static void
releaseme(void *arg)
{
	struct rumprunner *rr = arg;

	pthread_mutex_lock(&w_mtx);
	rumprun_done++;
	rr->rr_flags |= RUMPRUNNER_DONE;
	pthread_cond_broadcast(&w_cv);
	pthread_mutex_unlock(&w_mtx);
}



extern mainlike_fn rumprun_notmain;
extern mainlike_fn rumprun_main1;
extern mainlike_fn rumprun_main2;
extern mainlike_fn rumprun_main3;
extern mainlike_fn rumprun_main4;
extern mainlike_fn rumprun_main5;
extern mainlike_fn rumprun_main6;
extern mainlike_fn rumprun_main7;
extern mainlike_fn rumprun_main8;


typedef void (*initfini_fn)(void);
extern const initfini_fn __y1_init_array_start[1];
extern const initfini_fn __y1_init_array_end[1];
extern const initfini_fn __y2_init_array_start[1];
extern const initfini_fn __y2_init_array_end[1];
extern const initfini_fn __y3_init_array_start[1];
extern const initfini_fn __y3_init_array_end[1];
extern const initfini_fn __y4_init_array_start[1];
extern const initfini_fn __y4_init_array_end[1];
extern const initfini_fn __y5_init_array_start[1];
extern const initfini_fn __y5_init_array_end[1];
extern const initfini_fn __y6_init_array_start[1];
extern const initfini_fn __y6_init_array_end[1];
extern const initfini_fn __y7_init_array_start[1];
extern const initfini_fn __y7_init_array_end[1];
extern const initfini_fn __y8_init_array_start[1];
extern const initfini_fn __y8_init_array_end[1];

static void *
mainbouncer(void *arg)
{
	
	struct rumprunner *rr = arg;
	const char *progname = rr->rr_argv[0];
	int rv;

	rump_pub_lwproc_switch(rr->rr_lwp);

	pthread_cleanup_push(releaseme, rr);
	
	const initfini_fn *fn;
	if (rr->rr_mainfun == rumprun_main1) {
		for (fn = __y1_init_array_start; fn < __y1_init_array_end; fn++)
			(*fn)();
	} else if (rr->rr_mainfun == rumprun_main2) {
		for (fn = __y2_init_array_start; fn < __y2_init_array_end; fn++)
			(*fn)();
	} else if (rr->rr_mainfun == rumprun_main3) {
		for (fn = __y3_init_array_start; fn < __y3_init_array_end; fn++)
			(*fn)();
	} else if (rr->rr_mainfun == rumprun_main4) {
		for (fn = __y4_init_array_start; fn < __y4_init_array_end; fn++)
			(*fn)();
	} else if (rr->rr_mainfun == rumprun_main5) {
		for (fn = __y5_init_array_start; fn < __y5_init_array_end; fn++)
			(*fn)();
	} else if (rr->rr_mainfun == rumprun_main6) {
		for (fn = __y6_init_array_start; fn < __y6_init_array_end; fn++)
			(*fn)();
	} else if (rr->rr_mainfun == rumprun_main7) {
		for (fn = __y7_init_array_start; fn < __y7_init_array_end; fn++)
			(*fn)();
	} else if (rr->rr_mainfun == rumprun_main8) {
		for (fn = __y8_init_array_start; fn < __y8_init_array_end; fn++)
			(*fn)();
	}

	fprintf(stderr,"\n=== calling \"%s\" main() ===\n\n", progname);
	rv = rr->rr_mainfun(rr->rr_argc, rr->rr_argv);
	fflush(stdout);
	fprintf(stderr,"\n=== main() of \"%s\" returned %d ===\n",
	    progname, rv);

	pthread_cleanup_pop(1);

	/*
	 * XXX: missing _netbsd_userlevel_fini().  See comment in
	 * rumprun_boot()
	 */

	/* exit() calls rumprun_pub_lwproc_releaselwp() (via pthread_exit()) */
	exit(rv);
}

static void
setupproc(struct rumprunner *rr)
{
	static int pipein = -1;
	int pipefd[2], newpipein;
	const char *progname = rr->rr_argv[0];

	if (rump_pub_lwproc_curlwp() != NULL) {
		errx(1, "setupproc() needs support for non-implicit callers");
	}

	/* is the target output a pipe? */
	if (rr->rr_flags & RUMPRUN_EXEC_PIPE) {
		if (pipe(pipefd) == -1) {
			err(1, "cannot create pipe for %s", progname);
		}
		newpipein = pipefd[0];
	} else {
		newpipein = -1;
	}

	rump_pub_lwproc_rfork(RUMP_RFFDG);
	rr->rr_lwp = rump_pub_lwproc_curlwp();

	/* set output pipe to stdout if piping */
	if ((rr->rr_flags & RUMPRUN_EXEC_PIPE) && pipefd[1] != STDOUT_FILENO) {
		if (dup2(pipefd[1], STDOUT_FILENO) == -1)
			err(1, "dup2 stdout");
		close(pipefd[1]);
	}
	if (pipein != -1 && pipein != STDIN_FILENO) {
		if (dup2(pipein, STDIN_FILENO) == -1)
			err(1, "dup2 input");
		close(pipein);
	}

	rump_pub_lwproc_switch(NULL);

	/* pipe descriptors have been copied.  close them in parent */
	if (rr->rr_flags & RUMPRUN_EXEC_PIPE) {
		close(pipefd[1]);
	}
	if (pipein != -1) {
		close(pipein);
	}

	pipein = newpipein;
}

void *
rumprun(int flags, int (*mainfun)(int, char *[]), int argc, char *argv[])
{
	struct rumprunner *rr;

	rr = malloc(sizeof(*rr));

	/* XXX: should we deep copy argc? */
	rr->rr_mainfun = mainfun;
	rr->rr_argc = argc;
	rr->rr_argv = argv;
	rr->rr_flags = flags; /* XXX */

	setupproc(rr);

	if (pthread_create(&rr->rr_mainthread, NULL, mainbouncer, rr) != 0) {
		fprintf(stderr, "rumprun: running %s failed\n", argv[0]);
		free(rr);
		return NULL;
	}
	LIST_INSERT_HEAD(&rumprunners, rr, rr_entries);

	/* async launch? */
	if ((flags & (RUMPRUN_EXEC_BACKGROUND | RUMPRUN_EXEC_PIPE)) != 0) {
		return rr;
	}

	pthread_mutex_lock(&w_mtx);
	while ((rr->rr_flags & (RUMPRUNNER_DONE|RUMPRUNNER_DAEMON)) == 0) {
		pthread_cond_wait(&w_cv, &w_mtx);
	}
	pthread_mutex_unlock(&w_mtx);

	if (rr->rr_flags & RUMPRUNNER_DONE) {
		rumprun_wait(rr);
		rr = NULL;
	}
	return rr;
}

int
rumprun_wait(void *cookie)
{
	struct rumprunner *rr = cookie;
	void *retval;

	pthread_join(rr->rr_mainthread, &retval);
	LIST_REMOVE(rr, rr_entries);
	free(rr);

	assert(rumprun_done > 0);
	rumprun_done--;

	return (int)(intptr_t)retval;
}

void *
rumprun_get_finished(void)
{
	struct rumprunner *rr;

	if (LIST_EMPTY(&rumprunners))
		return NULL;

	pthread_mutex_lock(&w_mtx);
	while (rumprun_done == 0) {
		pthread_cond_wait(&w_cv, &w_mtx);
	}
	LIST_FOREACH(rr, &rumprunners, rr_entries) {
		if (rr->rr_flags & RUMPRUNNER_DONE) {
			break;
		}
	}
	pthread_mutex_unlock(&w_mtx);
	assert(rr);

	return rr;
}

/*
 * Detaches current program.  Must always be called from
 * the main thread of an application.  That's fine, since
 * given that the counterpart on a regular system (daemon()) forks,
 * it too must be called before threads are taken into use.
 *
 * It is expected that POSIX programs call this routine via daemon().
 */
void
rumprun_daemon(void)
{
	struct rumprunner *rr;

	LIST_FOREACH(rr, &rumprunners, rr_entries) {
		if (rr->rr_mainthread == pthread_self())
			break;
	}
	assert(rr);

	pthread_mutex_lock(&w_mtx);
	rr->rr_flags |= RUMPRUNNER_DAEMON;
	pthread_cond_broadcast(&w_cv);
	pthread_mutex_unlock(&w_mtx);
}

void __attribute__((noreturn))
rumprun_reboot(void)
{

	_netbsd_userlevel_fini();
	rump_sys_reboot(0, 0);

	bmk_platform_halt("reboot returned");
}
