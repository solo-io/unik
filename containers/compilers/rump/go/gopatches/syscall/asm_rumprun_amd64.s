// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO(rsc): Rewrite all nn(SP) references into name+(nn-8)(FP)
// so that go vet can check that they are correct.

#include "textflag.h"
#include "funcdata.h"

//
// System call support for AMD64, NetBSD
//

// func Syscall(trap int64, a1, a2, a3 int64) (r1, r2, err int64);
// func Syscall6(trap int64, a1, a2, a3, a4, a5, a6 int64) (r1, r2, err int64);
// Trap # in AX, args in DI SI DX, return in AX DX

TEXT	·Syscall(SB),NOSPLIT,$0-56
    JMP	runtime·ksyscall(SB)


TEXT	·Syscall6(SB),NOSPLIT,$0-80
	CALL	runtime·entersyscall(SB)
	MOVQ	8(SP), DI
	MOVQ	SP, SI
	ADDQ	$16, SI
	MOVQ	$0, DX		// dlen is ignored for local calls
	MOVQ	SP, CX
	ADDQ	$64, CX
	LEAQ	rump_syscall(SB), AX
	CALL	AX
	TESTQ	AX, AX
	JE	ok6
	MOVQ	AX, 80(SP)  	// errno
	CALL	runtime·exitsyscall(SB)
	RET
ok6:
	MOVQ	$0, 80(SP)	// errno
	CALL	runtime·exitsyscall(SB)
	RET

TEXT	·RawSyscall(SB),NOSPLIT,$0-56
	MOVQ	8(SP), DI
	MOVQ	SP, SI
	ADDQ	$16, SI
	MOVQ	$0, DX		// dlen is ignored for local calls
	MOVQ	SP, CX
	ADDQ	$40, CX
	LEAQ	rump_syscall(SB), AX
	CALL	AX
	TESTQ	AX, AX
	JE	ok1
	MOVQ	AX, 56(SP)	// errno
	RET
ok1:
	MOVQ	$0, 56(SP)	// errno
	RET

TEXT	·RawSyscall6(SB),NOSPLIT,$0-80
	MOVQ	8(SP), DI
	MOVQ	SP, SI
	ADDQ	$16, SI
	MOVQ	$0, DX		// dlen is ignored for local calls
	MOVQ	SP, CX
	ADDQ	$64, CX
	LEAQ	rump_syscall(SB), AX
	CALL	AX
	TESTQ	AX, AX
	JE	ok1
	MOVQ	AX, 80(SP)	// errno
	RET
ok1:
	MOVQ	$0, 80(SP)	// errno
	RET
