// +build darwin

#ifndef _GO_PROCESSDARWIN_H_INCLUDED
#define _GO_PROCESSDARWIN_H_INCLUDED

#include <ctype.h>
#include <errno.h>
#include <string.h>
#include <stdlib.h>
#include <sys/sysctl.h>

// This is declared in process_darwin.go
extern void go_darwin_append_proc(pid_t, pid_t, char *, int, char ***);

// This verifies if the current character under consideration is valid
// executable or argument character or not. There are cases seen where
// the arguments are not separated from the executable name by a NUL
// character but instead with a whitespace.
static inline int __isvalid(int c) {
	return c && !isspace(c);
}

// Loads the process table and calls the exported Go function to insert
// the data back into the Go space.
//
// This function is implemented in C because while it would technically
// be possible to do this all in Go, I didn't want to go spelunking through
// header files to get all the structures properly. It is much easier to just
// call it in C and be done with it.
static inline int darwinProcesses() {
    int err = 0;
    int i = 0, j = 0;
    static int name[] = { CTL_KERN, KERN_PROC, KERN_PROC_ALL, 0 };
    static int args[] = { CTL_KERN, 0, 0 };
    size_t length = 0;
    struct kinfo_proc *result = NULL;
    size_t resultCount = 0;

    // Get the length first
    err = sysctl(name, (sizeof(name) / sizeof(*name)) - 1,
            NULL, &length, NULL, 0);
    if (err != 0) {
        goto ERREXIT;
    }

    // Allocate the appropriate sized buffer to read the process list
    result = malloc(length);
    if (!result) {
        goto ERREXIT;
    }

    // Call sysctl again with our buffer to fill it with the process list
    err = sysctl(name, (sizeof(name) / sizeof(*name)) - 1,
            result, &length,
            NULL, 0);
    if (err != 0) {
        goto ERREXIT;
    }

    resultCount = length / sizeof(struct kinfo_proc);
    for (i = 0; i < resultCount; i++) {
        struct kinfo_proc *single = &result[i];
        pid_t pid = single->kp_proc.p_pid;
        char *p, *proc_argv, **argv;
        size_t count;
        int nargs, argmax, offset = 0;

        args[1] = KERN_ARGMAX;

        count = sizeof argmax;
        err = sysctl(args, 2, &argmax, &count, NULL, 0);
        if (err) {
            goto ERREXIT;
        }

        proc_argv = malloc(argmax);
        if (!proc_argv) {
            goto ERREXIT;
        }

        args[1] = KERN_PROCARGS2;
        args[2] = pid;

        count = argmax;
        err = sysctl(args, 3, proc_argv, &count, NULL, 0);
        if (err) {
            // We explicitly return if no valid arguments are found
            // for this command.
            if (errno == EINVAL) {
                // Mark errno as 0, as that is the value returned from
                // any calls made to a multi-valued cgo function.
                errno = 0;

                // No arguments were found and we inform our Go
                // counterpart for the processing it needs to skip.
                nargs = 0;

                // No cleanup is necessary for argv as nothing was
                // allocated for it.
                argv = NULL;
                goto RESULT;
            }

            goto ERREXIT;
        }

        // `proc_argv` contains the following data in order.
        //  1. count of arguments, includes the binary.
        //  2. complete executable path.
        //  3. name of the binary, delimited usually by a NUL or a whitespace.
        //  4. list of arguments, delimited usually by NUL or a whitespace.

        memcpy(&nargs, proc_argv, sizeof nargs);
        p = proc_argv + sizeof nargs;

        // `argv` will contain all the arguments of the process, starting with
        // the name of the binary.
        argv = malloc(sizeof (char *) * nargs);
        if (!argv) {
            goto ERREXIT;
        }

        // Save all the arguments in the argv that will be returned to Go.
        for (offset = 0; offset < nargs; offset++) {
            while (__isvalid(*p)) p++;
            while (!__isvalid(*p)) p++;

            argv[offset] = p;
        }

RESULT:
        go_darwin_append_proc(
                pid,
                single->kp_eproc.e_ppid,
                single->kp_proc.p_comm,
                nargs,
                &argv);

        free(proc_argv);

        if (argv != NULL) {
            free(argv);
        }
    }

ERREXIT:
    if (result != NULL) {
        free(result);
    }

    if (err != 0) {
        return errno;
    }
    return 0;
}

#endif
