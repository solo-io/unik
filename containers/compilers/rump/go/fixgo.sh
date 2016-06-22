# see https://github.com/deferpanic/gorump/issues/29
cat > runtime/string1.go <<EOF
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

//go:nosplit
func findnull(s *byte) int {
	if s == nil {
		return 0
	}
	p := (*[_MaxMem/2 - 1]byte)(unsafe.Pointer(s))
	l := 0
  for l < len(p) && p[l] != 0 {
          l++
  }
  if l == len(p) {
    return 0
  }
	return l
}

func findnullw(s *uint16) int {
	if s == nil {
		return 0
	}
	p := (*[_MaxMem/2/2 - 1]uint16)(unsafe.Pointer(s))
	l := 0
  for l < len(p) && p[l] != 0 {
          l++
  }
  if l == len(p) {
    return 0
  }
	return l
}

var maxstring uintptr = 256 // a hint for print

//go:nosplit
func gostringnocopy(str *byte) string {
	ss := stringStruct{str: unsafe.Pointer(str), len: findnull(str)}
	s := *(*string)(unsafe.Pointer(&ss))
	for {
		ms := maxstring
		if uintptr(len(s)) <= ms || casuintptr(&maxstring, ms, uintptr(len(s))) {
			break
		}
	}
	return s
}

func gostringw(strw *uint16) string {
	var buf [8]byte
	str := (*[_MaxMem/2/2 - 1]uint16)(unsafe.Pointer(strw))
	n1 := 0
	for i := 0; str[i] != 0; i++ {
		n1 += runetochar(buf[:], rune(str[i]))
	}
	s, b := rawstring(n1 + 4)
	n2 := 0
	for i := 0; str[i] != 0; i++ {
		// check for race
		if n2 >= n1 {
			break
		}
		n2 += runetochar(b[n2:], rune(str[i]))
	}
	b[n2] = 0 // for luck
	return s[:n2]
}
EOF

cat > runtime/os1_rumprun.go <<EOF
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

const (
	_ESRCH     = 3
	_ETIMEDOUT = 60

	// From NetBSD's <sys/time.h>
	_CLOCK_REALTIME  = 0
	_CLOCK_VIRTUAL   = 1
	_CLOCK_PROF      = 2
	_CLOCK_MONOTONIC = 3
)

var sigset_all = sigset{[4]uint32{^uint32(0), ^uint32(0), ^uint32(0), ^uint32(0)}}

// From NetBSD's <sys/sysctl.h>
const (
	_CTL_HW  = 6
	_HW_NCPU = 3
)

func getncpu() int32 {
	mib := [2]uint32{_CTL_HW, _HW_NCPU}
	out := uint32(0)
	nout := unsafe.Sizeof(out)
	ret := sysctl(&mib[0], 2, (*byte)(unsafe.Pointer(&out)), &nout, nil, 0)
	if ret >= 0 {
		return int32(out)
	}
	return 1
}

//go:nosplit
func semacreate() uintptr {
	return 1
}

//go:nosplit
func semasleep(ns int64) int32 {
	_g_ := getg()

	// Compute sleep deadline.
	var tsp *timespec
	if ns >= 0 {
		var ts timespec
		var nsec int32
		ns += nanotime()
		ts.set_sec(timediv(ns, 1000000000, &nsec))
		ts.set_nsec(nsec)
		tsp = &ts
	}

	for {
		v := atomicload(&_g_.m.waitsemacount)
		if v > 0 {
			if cas(&_g_.m.waitsemacount, v, v-1) {
				return 0 // semaphore acquired
			}
			continue
		}

		// Sleep until unparked by semawakeup or timeout.
		ret := lwp_park(tsp, 0, unsafe.Pointer(&_g_.m.waitsemacount), nil)
		if ret == _ETIMEDOUT || ret == -1 {
			return -1
		}
	}
}

//go:nosplit
func semawakeup(mp *m) {
	xadd(&mp.waitsemacount, 1)
	// From NetBSD's _lwp_unpark(2) manual:
	// "If the target LWP is not currently waiting, it will return
	// immediately upon the next call to _lwp_park()."
	ret := lwp_unpark(int32(mp.procid), unsafe.Pointer(&mp.waitsemacount))
	if ret != 0 && ret != _ESRCH {
		// semawakeup can be called on signal stack.
		systemstack(func() {
			print("thrwakeup addr=", &mp.waitsemacount, " sem=", mp.waitsemacount, " ret=", ret, "\n")
		})
	}
}

// May run with m.p==nil, so write barriers are not allowed.
//go:nowritebarrier
func newosproc(mp *m, stk unsafe.Pointer) {
	if false {
		print("newosproc stk=", stk, " m=", mp, " g=", mp.g0, " id=", mp.id, "/", int32(mp.tls[0]), " ostk=", &mp, "\n")
	}

	mp.tls[0] = uintptr(mp.id) // so 386 asm can find it

	var uc ucontextt
	getcontext(unsafe.Pointer(&uc))

	uc.uc_flags = _UC_SIGMASK | _UC_CPU
	uc.uc_link = nil
	uc.uc_sigmask = sigset_all

	lwp_mcontext_init(&uc.uc_mcontext, stk, mp, mp.g0, funcPC(mstart))

	ret := lwp_create(unsafe.Pointer(&uc), 0, unsafe.Pointer(&mp.procid))
	if ret < 0 {
		print("runtime: failed to create new OS thread (have ", mcount()-1, " already; errno=", -ret, ")\n")
		throw("runtime.newosproc")
	}
}

func osinit() {
	ncpu = getncpu()
}

var urandom_dev = []byte("/dev/urandom\x00")

//go:nosplit
func getRandomData(r []byte) {
	fd := open(&urandom_dev[0], 0 /* O_RDONLY */, 0)
	n := read(fd, unsafe.Pointer(&r[0]), int32(len(r)))
	closefd(fd)
	extendRandom(r, int(n))
}

func goenvs() {
	goenvs_unix()
}

// Called to initialize a new m (including the bootstrap m).
// Called on the parent thread (main thread in case of bootstrap), can allocate memory.
func mpreinit(mp *m) {
	mp.gsignal = malg(32 * 1024)
	mp.gsignal.m = mp
}

func msigsave(mp *m) {
}

// Called to initialize a new m (including the bootstrap m).
// Called on the new thread, can not allocate memory.
func minit() {
	_g_ := getg()
	_g_.m.procid = uint64(lwp_self())
}

// Called from dropm to undo the effect of an minit.
func unminit() {
}

func memlimit() uintptr {
	return 0
}

func sigtramp()

type sigactiont struct {
	sa_sigaction uintptr
	sa_mask      sigset
	sa_flags     int32
}

func setsig(i int32, fn uintptr, restart bool) {
}

func setsigstack(i int32) {
	throw("setsigstack")
}

func getsig(i int32) uintptr {
	return 0
}

func signalstack(s *stack) {
}

func updatesigmask(m sigmask) {
}

func unblocksig(sig int32) {
}

EOF

# applying:
# https://github.com/golang/go/commit/0de59c27ebda76d3e5121888acca304e49009266
# https://github.com/golang/go/commit/f9357cdec1235168590283abe027e501633755a5
# https://github.com/golang/go/commit/97dc591534e580ede4c6145456dadf1e009f1bea
# see https://github.com/golang/go/issues/13143

cat > runtime/malloc.go <<EOF
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Memory allocator, based on tcmalloc.
// http://goog-perftools.sourceforge.net/doc/tcmalloc.html

// The main allocator works in runs of pages.
// Small allocation sizes (up to and including 32 kB) are
// rounded to one of about 100 size classes, each of which
// has its own free list of objects of exactly that size.
// Any free page of memory can be split into a set of objects
// of one size class, which are then managed using free list
// allocators.
//
// The allocator's data structures are:
//
//	FixAlloc: a free-list allocator for fixed-size objects,
//		used to manage storage used by the allocator.
//	MHeap: the malloc heap, managed at page (4096-byte) granularity.
//	MSpan: a run of pages managed by the MHeap.
//	MCentral: a shared free list for a given size class.
//	MCache: a per-thread (in Go, per-P) cache for small objects.
//	MStats: allocation statistics.
//
// Allocating a small object proceeds up a hierarchy of caches:
//
//	1. Round the size up to one of the small size classes
//	   and look in the corresponding MCache free list.
//	   If the list is not empty, allocate an object from it.
//	   This can all be done without acquiring a lock.
//
//	2. If the MCache free list is empty, replenish it by
//	   taking a bunch of objects from the MCentral free list.
//	   Moving a bunch amortizes the cost of acquiring the MCentral lock.
//
//	3. If the MCentral free list is empty, replenish it by
//	   allocating a run of pages from the MHeap and then
//	   chopping that memory into objects of the given size.
//	   Allocating many objects amortizes the cost of locking
//	   the heap.
//
//	4. If the MHeap is empty or has no page runs large enough,
//	   allocate a new group of pages (at least 1MB) from the
//	   operating system.  Allocating a large run of pages
//	   amortizes the cost of talking to the operating system.
//
// Freeing a small object proceeds up the same hierarchy:
//
//	1. Look up the size class for the object and add it to
//	   the MCache free list.
//
//	2. If the MCache free list is too long or the MCache has
//	   too much memory, return some to the MCentral free lists.
//
//	3. If all the objects in a given span have returned to
//	   the MCentral list, return that span to the page heap.
//
//	4. If the heap has too much memory, return some to the
//	   operating system.
//
//	TODO(rsc): Step 4 is not implemented.
//
// Allocating and freeing a large object uses the page heap
// directly, bypassing the MCache and MCentral free lists.
//
// The small objects on the MCache and MCentral free lists
// may or may not be zeroed.  They are zeroed if and only if
// the second word of the object is zero.  A span in the
// page heap is zeroed unless s->needzero is set. When a span
// is allocated to break into small objects, it is zeroed if needed
// and s->needzero is set. There are two main benefits to delaying the
// zeroing this way:
//
//	1. stack frames allocated from the small object lists
//	   or the page heap can avoid zeroing altogether.
//	2. the cost of zeroing when reusing a small object is
//	   charged to the mutator, not the garbage collector.
//
// This code was written with an eye toward translating to Go
// in the future.  Methods have the form Type_Method(Type *t, ...).

package runtime

import "unsafe"

const (
	debugMalloc = false

	flagNoScan = _FlagNoScan
	flagNoZero = _FlagNoZero

	maxTinySize   = _TinySize
	tinySizeClass = _TinySizeClass
	maxSmallSize  = _MaxSmallSize

	pageShift = _PageShift
	pageSize  = _PageSize
	pageMask  = _PageMask

	mSpanInUse = _MSpanInUse

	concurrentSweep = _ConcurrentSweep
)

const (
	_PageShift = 13
	_PageSize  = 1 << _PageShift
	_PageMask  = _PageSize - 1
)

const (
	// _64bit = 1 on 64-bit systems, 0 on 32-bit systems
	_64bit = 1 << (^uintptr(0) >> 63) / 2

	// Computed constant.  The definition of MaxSmallSize and the
	// algorithm in msize.go produces some number of different allocation
	// size classes.  NumSizeClasses is that number.  It's needed here
	// because there are static arrays of this length; when msize runs its
	// size choosing algorithm it double-checks that NumSizeClasses agrees.
	_NumSizeClasses = 67

	// Tunable constants.
	_MaxSmallSize = 32 << 10

	// Tiny allocator parameters, see "Tiny allocator" comment in malloc.go.
	_TinySize      = 16
	_TinySizeClass = 2

	_FixAllocChunk  = 16 << 10               // Chunk size for FixAlloc
	_MaxMHeapList   = 1 << (20 - _PageShift) // Maximum page length for fixed-size list in MHeap.
	_HeapAllocChunk = 1 << 20                // Chunk size for heap growth

	// Per-P, per order stack segment cache size.
	_StackCacheSize = 32 * 1024

	// Number of orders that get caching.  Order 0 is FixedStack
	// and each successive order is twice as large.
	// We want to cache 2KB, 4KB, 8KB, and 16KB stacks.  Larger stacks
	// will be allocated directly.
	// Since FixedStack is different on different systems, we
	// must vary NumStackOrders to keep the same maximum cached size.
	//   OS               | FixedStack | NumStackOrders
	//   -----------------+------------+---------------
	//   linux/darwin/bsd | 2KB        | 4
	//   windows/32       | 4KB        | 3
	//   windows/64       | 8KB        | 2
	//   plan9            | 4KB        | 3
	_NumStackOrders = 4 - ptrSize/4*goos_windows - 1*goos_plan9

	// Number of bits in page to span calculations (4k pages).
	// On Windows 64-bit we limit the arena to 32GB or 35 bits.
	// Windows counts memory used by page table into committed memory
	// of the process, so we can't reserve too much memory.
	// See https://golang.org/issue/5402 and https://golang.org/issue/5236.
	// On other 64-bit platforms, we limit the arena to 512GB, or 39 bits.
	// On 32-bit, we don't bother limiting anything, so we use the full 32-bit address.
	// On Darwin/arm64, we cannot reserve more than ~5GB of virtual memory,
	// but as most devices have less than 4GB of physical memory anyway, we
	// try to be conservative here, and only ask for a 2GB heap.

	// XXX: use goos_netbsd really as good_rumprun
	_MHeapMap_TotalBits = (_64bit*goos_windows)*35 + (1-goos_rumprun)*(_64bit*(1-goos_windows)*(1-goos_darwin*goarch_arm64))*39 + goos_darwin*goarch_arm64*31 + (1-_64bit)*32*(1-goos_rumprun) + 25*goos_rumprun

	_MHeapMap_Bits = _MHeapMap_TotalBits - _PageShift

	_MaxMem = uintptr(1<<_MHeapMap_TotalBits - 1)

	// Max number of threads to run garbage collection.
	// 2, 3, and 4 are all plausible maximums depending
	// on the hardware details of the machine.  The garbage
	// collector scales well to 32 cpus.
	_MaxGcproc = 32
)

// Page number (address>>pageShift)
type pageID uintptr

const _MaxArena32 = 2 << 30

// OS-defined helpers:
//
// sysAlloc obtains a large chunk of zeroed memory from the
// operating system, typically on the order of a hundred kilobytes
// or a megabyte.
// NOTE: sysAlloc returns OS-aligned memory, but the heap allocator
// may use larger alignment, so the caller must be careful to realign the
// memory obtained by sysAlloc.
//
// SysUnused notifies the operating system that the contents
// of the memory region are no longer needed and can be reused
// for other purposes.
// SysUsed notifies the operating system that the contents
// of the memory region are needed again.
//
// SysFree returns it unconditionally; this is only used if
// an out-of-memory error has been detected midway through
// an allocation.  It is okay if SysFree is a no-op.
//
// SysReserve reserves address space without allocating memory.
// If the pointer passed to it is non-nil, the caller wants the
// reservation there, but SysReserve can still choose another
// location if that one is unavailable.  On some systems and in some
// cases SysReserve will simply check that the address space is
// available and not actually reserve it.  If SysReserve returns
// non-nil, it sets *reserved to true if the address space is
// reserved, false if it has merely been checked.
// NOTE: SysReserve returns OS-aligned memory, but the heap allocator
// may use larger alignment, so the caller must be careful to realign the
// memory obtained by sysAlloc.
//
// SysMap maps previously reserved address space for use.
// The reserved argument is true if the address space was really
// reserved, not merely checked.
//
// SysFault marks a (already sysAlloc'd) region to fault
// if accessed.  Used only for debugging the runtime.

func mallocinit() {
	initSizes()

	if class_to_size[_TinySizeClass] != _TinySize {
		throw("bad TinySizeClass")
	}

	var p, bitmapSize, spansSize, pSize, limit uintptr
	var reserved bool

	// limit = runtime.memlimit();
	// See https://golang.org/issue/5049
	// TODO(rsc): Fix after 1.1.
	limit = 0

	// Set up the allocation arena, a contiguous area of memory where
	// allocated data will be found.  The arena begins with a bitmap large
	// enough to hold 4 bits per allocated word.
	if ptrSize == 8 && (limit == 0 || limit > 1<<30) {
		// On a 64-bit machine, allocate from a single contiguous reservation.
		// 512 GB (MaxMem) should be big enough for now.
		//
		// The code will work with the reservation at any address, but ask
		// SysReserve to use 0x0000XXc000000000 if possible (XX=00...7f).
		// Allocating a 512 GB region takes away 39 bits, and the amd64
		// doesn't let us choose the top 17 bits, so that leaves the 9 bits
		// in the middle of 0x00c0 for us to choose.  Choosing 0x00c0 means
		// that the valid memory addresses will begin 0x00c0, 0x00c1, ..., 0x00df.
		// In little-endian, that's c0 00, c1 00, ..., df 00. None of those are valid
		// UTF-8 sequences, and they are otherwise as far away from
		// ff (likely a common byte) as possible.  If that fails, we try other 0xXXc0
		// addresses.  An earlier attempt to use 0x11f8 caused out of memory errors
		// on OS X during thread allocations.  0x00c0 causes conflicts with
		// AddressSanitizer which reserves all memory up to 0x0100.
		// These choices are both for debuggability and to reduce the
		// odds of a conservative garbage collector (as is still used in gccgo)
		// not collecting memory because some non-pointer block of memory
		// had a bit pattern that matched a memory address.
		//
		// Actually we reserve 544 GB (because the bitmap ends up being 32 GB)
		// but it hardly matters: e0 00 is not valid UTF-8 either.
		//
		// If this fails we fall back to the 32 bit memory mechanism
		//
		// However, on arm64, we ignore all this advice above and slam the
		// allocation at 0x40 << 32 because when using 4k pages with 3-level
		// translation buffers, the user address space is limited to 39 bits
		// On darwin/arm64, the address space is even smaller.
		arenaSize := round(_MaxMem, _PageSize)
		bitmapSize = arenaSize / (ptrSize * 8 / 4)
		spansSize = arenaSize / _PageSize * ptrSize
		spansSize = round(spansSize, _PageSize)
		for i := 0; i <= 0x7f; i++ {
			switch {
			case GOARCH == "arm64" && GOOS == "darwin":
				p = uintptr(i)<<40 | uintptrMask&(0x0013<<28)
			case GOARCH == "arm64":
				p = uintptr(i)<<40 | uintptrMask&(0x0040<<32)
			default:
				p = uintptr(i)<<40 | uintptrMask&(0x00c0<<32)
			}
			pSize = bitmapSize + spansSize + arenaSize + _PageSize
			p = uintptr(sysReserve(unsafe.Pointer(p), pSize, &reserved))
			if p != 0 {
				break
			}
		}
	}

	if p == 0 {
		// On a 32-bit machine, we can't typically get away
		// with a giant virtual address space reservation.
		// Instead we map the memory information bitmap
		// immediately after the data segment, large enough
		// to handle another 2GB of mappings (256 MB),
		// along with a reservation for an initial arena.
		// When that gets used up, we'll start asking the kernel
		// for any memory anywhere and hope it's in the 2GB
		// following the bitmap (presumably the executable begins
		// near the bottom of memory, so we'll have to use up
		// most of memory before the kernel resorts to giving out
		// memory before the beginning of the text segment).
		//
		// Alternatively we could reserve 512 MB bitmap, enough
		// for 4GB of mappings, and then accept any memory the
		// kernel threw at us, but normally that's a waste of 512 MB
		// of address space, which is probably too much in a 32-bit world.

		// If we fail to allocate, try again with a smaller arena.
		// This is necessary on Android L where we share a process
		// with ART, which reserves virtual memory aggressively.
		arenaSizes := []uintptr{
			512 << 20,
			256 << 20,
			128 << 20,
		}

		for _, arenaSize := range arenaSizes {
			bitmapSize = _MaxArena32 / (ptrSize * 8 / 4)
			spansSize = _MaxArena32 / _PageSize * ptrSize
			if limit > 0 && arenaSize+bitmapSize+spansSize > limit {
				bitmapSize = (limit / 9) &^ ((1 << _PageShift) - 1)
				arenaSize = bitmapSize * 8
				spansSize = arenaSize / _PageSize * ptrSize
			}
			spansSize = round(spansSize, _PageSize)

			// SysReserve treats the address we ask for, end, as a hint,
			// not as an absolute requirement.  If we ask for the end
			// of the data segment but the operating system requires
			// a little more space before we can start allocating, it will
			// give out a slightly higher pointer.  Except QEMU, which
			// is buggy, as usual: it won't adjust the pointer upward.
			// So adjust it upward a little bit ourselves: 1/4 MB to get
			// away from the running binary image and then round up
			// to a MB boundary.
			p = round(firstmoduledata.end+(1<<18), 1<<20)
			pSize = bitmapSize + spansSize + arenaSize + _PageSize
			p = uintptr(sysReserve(unsafe.Pointer(p), pSize, &reserved))
			if p != 0 {
				break
			}
		}
		if p == 0 {
			throw("runtime: cannot reserve arena virtual address space")
		}
	}

	// PageSize can be larger than OS definition of page size,
	// so SysReserve can give us a PageSize-unaligned pointer.
	// To overcome this we ask for PageSize more and round up the pointer.
	p1 := round(p, _PageSize)

	mheap_.spans = (**mspan)(unsafe.Pointer(p1))
	mheap_.bitmap = p1 + spansSize
	mheap_.arena_start = p1 + (spansSize + bitmapSize)
	mheap_.arena_used = mheap_.arena_start
	mheap_.arena_end = p + pSize
	mheap_.arena_reserved = reserved

	if mheap_.arena_start&(_PageSize-1) != 0 {
		println("bad pagesize", hex(p), hex(p1), hex(spansSize), hex(bitmapSize), hex(_PageSize), "start", hex(mheap_.arena_start))
		throw("misrounded allocation in mallocinit")
	}

	// Initialize the rest of the allocator.
	mHeap_Init(&mheap_, spansSize)
	_g_ := getg()
	_g_.m.mcache = allocmcache()
}

// sysReserveHigh reserves space somewhere high in the address space.
// sysReserve doesn't actually reserve the full amount requested on
// 64-bit systems, because of problems with ulimit. Instead it checks
// that it can get the first 64 kB and assumes it can grab the rest as
// needed. This doesn't work well with the "let the kernel pick an address"
// mode, so don't do that. Pick a high address instead.
func sysReserveHigh(n uintptr, reserved *bool) unsafe.Pointer {
	if ptrSize == 4 {
		return sysReserve(nil, n, reserved)
	}

	for i := 0; i <= 0x7f; i++ {
		p := uintptr(i)<<40 | uintptrMask&(0x00c0<<32)
		*reserved = false
		p = uintptr(sysReserve(unsafe.Pointer(p), n, reserved))
		if p != 0 {
			return unsafe.Pointer(p)
		}
	}

	return sysReserve(nil, n, reserved)
}

func mHeap_SysAlloc(h *mheap, n uintptr) unsafe.Pointer {
	if n > uintptr(h.arena_end)-uintptr(h.arena_used) {
		// We are in 32-bit mode, maybe we didn't use all possible address space yet.
		// Reserve some more space.
		p_size := round(n+_PageSize, 256<<20)
		new_end := h.arena_end + p_size // Careful: can overflow
		if h.arena_end <= new_end && new_end <= h.arena_start+_MaxArena32 {
			// TODO: It would be bad if part of the arena
			// is reserved and part is not.
			var reserved bool
			p := uintptr(sysReserve((unsafe.Pointer)(h.arena_end), p_size, &reserved))
			if p == h.arena_end {
				h.arena_end = new_end
				h.arena_reserved = reserved
			} else if h.arena_start <= p && p+p_size <= h.arena_start+_MaxArena32 {
				// Keep everything page-aligned.
				// Our pages are bigger than hardware pages.
				h.arena_end = p + p_size
				used := p + (-uintptr(p) & (_PageSize - 1))
				mHeap_MapBits(h, used)
				mHeap_MapSpans(h, used)
				h.arena_used = used
				h.arena_reserved = reserved
			} else {
				stat := uint64(p_size)
				sysFree((unsafe.Pointer)(p), p_size, &stat)
			}
		}
	}

	if n <= uintptr(h.arena_end)-uintptr(h.arena_used) {
		// Keep taking from our reservation.
		p := h.arena_used
		sysMap((unsafe.Pointer)(p), n, h.arena_reserved, &memstats.heap_sys)
		mHeap_MapBits(h, p+n)
		mHeap_MapSpans(h, p+n)
		h.arena_used = p + n
		if raceenabled {
			racemapshadow((unsafe.Pointer)(p), n)
		}

		if uintptr(p)&(_PageSize-1) != 0 {
			throw("misrounded allocation in MHeap_SysAlloc")
		}
		return (unsafe.Pointer)(p)
	}

	// If using 64-bit, our reservation is all we have.
	if uintptr(h.arena_end)-uintptr(h.arena_start) >= _MaxArena32 {
		return nil
	}

	// On 32-bit, once the reservation is gone we can
	// try to get memory at a location chosen by the OS
	// and hope that it is in the range we allocated bitmap for.
	p_size := round(n, _PageSize) + _PageSize
	p := uintptr(sysAlloc(p_size, &memstats.heap_sys))
	if p == 0 {
		return nil
	}

	if p < h.arena_start || uintptr(p)+p_size-uintptr(h.arena_start) >= _MaxArena32 {
		print("runtime: memory allocated by OS (", p, ") not in usable range [", hex(h.arena_start), ",", hex(h.arena_start+_MaxArena32), ")\n")
		sysFree((unsafe.Pointer)(p), p_size, &memstats.heap_sys)
		return nil
	}

	p_end := p + p_size
	p += -p & (_PageSize - 1)
	if uintptr(p)+n > uintptr(h.arena_used) {
		mHeap_MapBits(h, p+n)
		mHeap_MapSpans(h, p+n)
		h.arena_used = p + n
		if p_end > h.arena_end {
			h.arena_end = p_end
		}
		if raceenabled {
			racemapshadow((unsafe.Pointer)(p), n)
		}
	}

	if uintptr(p)&(_PageSize-1) != 0 {
		throw("misrounded allocation in MHeap_SysAlloc")
	}
	return (unsafe.Pointer)(p)
}

// base address for all 0-byte allocations
var zerobase uintptr

const (
	// flags to malloc
	_FlagNoScan = 1 << 0 // GC doesn't have to scan object
	_FlagNoZero = 1 << 1 // don't zero memory
)

// Allocate an object of size bytes.
// Small objects are allocated from the per-P cache's free lists.
// Large objects (> 32 kB) are allocated straight from the heap.
func mallocgc(size uintptr, typ *_type, flags uint32) unsafe.Pointer {
	if gcphase == _GCmarktermination {
		throw("mallocgc called with gcphase == _GCmarktermination")
	}

	if size == 0 {
		return unsafe.Pointer(&zerobase)
	}

	if flags&flagNoScan == 0 && typ == nil {
		throw("malloc missing type")
	}

	if debug.sbrk != 0 {
		align := uintptr(16)
		if typ != nil {
			align = uintptr(typ.align)
		}
		return persistentalloc(size, align, &memstats.other_sys)
	}

	// Set mp.mallocing to keep from being preempted by GC.
	mp := acquirem()
	if mp.mallocing != 0 {
		throw("malloc deadlock")
	}
	if mp.gsignal == getg() {
		throw("malloc during signal")
	}
	mp.mallocing = 1

	shouldhelpgc := false
	dataSize := size
	c := gomcache()
	var s *mspan
	var x unsafe.Pointer
	if size <= maxSmallSize {
		if flags&flagNoScan != 0 && size < maxTinySize {
			// Tiny allocator.
			//
			// Tiny allocator combines several tiny allocation requests
			// into a single memory block. The resulting memory block
			// is freed when all subobjects are unreachable. The subobjects
			// must be FlagNoScan (don't have pointers), this ensures that
			// the amount of potentially wasted memory is bounded.
			//
			// Size of the memory block used for combining (maxTinySize) is tunable.
			// Current setting is 16 bytes, which relates to 2x worst case memory
			// wastage (when all but one subobjects are unreachable).
			// 8 bytes would result in no wastage at all, but provides less
			// opportunities for combining.
			// 32 bytes provides more opportunities for combining,
			// but can lead to 4x worst case wastage.
			// The best case winning is 8x regardless of block size.
			//
			// Objects obtained from tiny allocator must not be freed explicitly.
			// So when an object will be freed explicitly, we ensure that
			// its size >= maxTinySize.
			//
			// SetFinalizer has a special case for objects potentially coming
			// from tiny allocator, it such case it allows to set finalizers
			// for an inner byte of a memory block.
			//
			// The main targets of tiny allocator are small strings and
			// standalone escaping variables. On a json benchmark
			// the allocator reduces number of allocations by ~12% and
			// reduces heap size by ~20%.
			off := c.tinyoffset
			// Align tiny pointer for required (conservative) alignment.
			if size&7 == 0 {
				off = round(off, 8)
			} else if size&3 == 0 {
				off = round(off, 4)
			} else if size&1 == 0 {
				off = round(off, 2)
			}
			if off+size <= maxTinySize && c.tiny != nil {
				// The object fits into existing tiny block.
				x = add(c.tiny, off)
				c.tinyoffset = off + size
				c.local_tinyallocs++
				mp.mallocing = 0
				releasem(mp)
				return x
			}
			// Allocate a new maxTinySize block.
			s = c.alloc[tinySizeClass]
			v := s.freelist
			if v.ptr() == nil {
				systemstack(func() {
					mCache_Refill(c, tinySizeClass)
				})
				shouldhelpgc = true
				s = c.alloc[tinySizeClass]
				v = s.freelist
			}
			s.freelist = v.ptr().next
			s.ref++
			// prefetchnta offers best performance, see change list message.
			prefetchnta(uintptr(v.ptr().next))
			x = unsafe.Pointer(v)
			(*[2]uint64)(x)[0] = 0
			(*[2]uint64)(x)[1] = 0
			// See if we need to replace the existing tiny block with the new one
			// based on amount of remaining free space.
			if size < c.tinyoffset {
				c.tiny = x
				c.tinyoffset = size
			}
			size = maxTinySize
		} else {
			var sizeclass int8
			if size <= 1024-8 {
				sizeclass = size_to_class8[(size+7)>>3]
			} else {
				sizeclass = size_to_class128[(size-1024+127)>>7]
			}
			size = uintptr(class_to_size[sizeclass])
			s = c.alloc[sizeclass]
			v := s.freelist
			if v.ptr() == nil {
				systemstack(func() {
					mCache_Refill(c, int32(sizeclass))
				})
				shouldhelpgc = true
				s = c.alloc[sizeclass]
				v = s.freelist
			}
			s.freelist = v.ptr().next
			s.ref++
			// prefetchnta offers best performance, see change list message.
			prefetchnta(uintptr(v.ptr().next))
			x = unsafe.Pointer(v)
			if flags&flagNoZero == 0 {
				v.ptr().next = 0
				if size > 2*ptrSize && ((*[2]uintptr)(x))[1] != 0 {
					memclr(unsafe.Pointer(v), size)
				}
			}
		}
		c.local_cachealloc += size
	} else {
		var s *mspan
		shouldhelpgc = true
		systemstack(func() {
			s = largeAlloc(size, uint32(flags))
		})
		x = unsafe.Pointer(uintptr(s.start << pageShift))
		size = uintptr(s.elemsize)
	}

	if flags&flagNoScan != 0 {
		// All objects are pre-marked as noscan. Nothing to do.
	} else {
		// If allocating a defer+arg block, now that we've picked a malloc size
		// large enough to hold everything, cut the "asked for" size down to
		// just the defer header, so that the GC bitmap will record the arg block
		// as containing nothing at all (as if it were unused space at the end of
		// a malloc block caused by size rounding).
		// The defer arg areas are scanned as part of scanstack.
		if typ == deferType {
			dataSize = unsafe.Sizeof(_defer{})
		}
		heapBitsSetType(uintptr(x), size, dataSize, typ)
		if dataSize > typ.size {
			// Array allocation. If there are any
			// pointers, GC has to scan to the last
			// element.
			if typ.ptrdata != 0 {
				c.local_scan += dataSize - typ.size + typ.ptrdata
			}
		} else {
			c.local_scan += typ.ptrdata
		}

		// Ensure that the stores above that initialize x to
		// type-safe memory and set the heap bits occur before
		// the caller can make x observable to the garbage
		// collector. Otherwise, on weakly ordered machines,
		// the garbage collector could follow a pointer to x,
		// but see uninitialized memory or stale heap bits.
		publicationBarrier()
	}

	// GCmarkterminate allocates black
	// All slots hold nil so no scanning is needed.
	// This may be racing with GC so do it atomically if there can be
	// a race marking the bit.
	if gcphase == _GCmarktermination || gcBlackenPromptly {
		systemstack(func() {
			gcmarknewobject_m(uintptr(x), size)
		})
	}

	if raceenabled {
		racemalloc(x, size)
	}

	mp.mallocing = 0
	releasem(mp)

	if debug.allocfreetrace != 0 {
		tracealloc(x, size, typ)
	}

	if rate := MemProfileRate; rate > 0 {
		if size < uintptr(rate) && int32(size) < c.next_sample {
			c.next_sample -= int32(size)
		} else {
			mp := acquirem()
			profilealloc(mp, x, size)
			releasem(mp)
		}
	}

	if shouldhelpgc && shouldtriggergc() {
		startGC(gcBackgroundMode, false)
	} else if gcBlackenEnabled != 0 {
		// Assist garbage collector. We delay this until the
		// epilogue so that it doesn't interfere with the
		// inner working of malloc such as mcache refills that
		// might happen while doing the gcAssistAlloc.
		gcAssistAlloc(size, shouldhelpgc)
	} else if shouldhelpgc && bggc.working != 0 {
		// The GC is starting up or shutting down, so we can't
		// assist, but we also can't allocate unabated. Slow
		// down this G's allocation and help the GC stay
		// scheduled by yielding.
		//
		// TODO: This is a workaround. Either help the GC make
		// the transition or block.
		gp := getg()
		if gp != gp.m.g0 && gp.m.locks == 0 && gp.m.preemptoff == "" {
			Gosched()
		}
	}

	return x
}

func largeAlloc(size uintptr, flag uint32) *mspan {
	// print("largeAlloc size=", size, "\n")

	if size+_PageSize < size {
		throw("out of memory")
	}
	npages := size >> _PageShift
	if size&_PageMask != 0 {
		npages++
	}

	// Deduct credit for this span allocation and sweep if
	// necessary. mHeap_Alloc will also sweep npages, so this only
	// pays the debt down to npage pages.
	deductSweepCredit(npages*_PageSize, npages)

	s := mHeap_Alloc(&mheap_, npages, 0, true, flag&_FlagNoZero == 0)
	if s == nil {
		throw("out of memory")
	}
	s.limit = uintptr(s.start)<<_PageShift + size
	heapBitsForSpan(s.base()).initSpan(s.layout())
	return s
}

// implementation of new builtin
func newobject(typ *_type) unsafe.Pointer {
	flags := uint32(0)
	if typ.kind&kindNoPointers != 0 {
		flags |= flagNoScan
	}
	return mallocgc(uintptr(typ.size), typ, flags)
}

//go:linkname reflect_unsafe_New reflect.unsafe_New
func reflect_unsafe_New(typ *_type) unsafe.Pointer {
	return newobject(typ)
}

// implementation of make builtin for slices
func newarray(typ *_type, n uintptr) unsafe.Pointer {
	flags := uint32(0)
	if typ.kind&kindNoPointers != 0 {
		flags |= flagNoScan
	}
	if int(n) < 0 || (typ.size > 0 && n > _MaxMem/uintptr(typ.size)) {
		panic("runtime: allocation size out of range")
	}
	return mallocgc(uintptr(typ.size)*n, typ, flags)
}

//go:linkname reflect_unsafe_NewArray reflect.unsafe_NewArray
func reflect_unsafe_NewArray(typ *_type, n uintptr) unsafe.Pointer {
	return newarray(typ, n)
}

// rawmem returns a chunk of pointerless memory.  It is
// not zeroed.
func rawmem(size uintptr) unsafe.Pointer {
	return mallocgc(size, nil, flagNoScan|flagNoZero)
}

func profilealloc(mp *m, x unsafe.Pointer, size uintptr) {
	c := mp.mcache
	rate := MemProfileRate
	if size < uintptr(rate) {
		// pick next profile time
		// If you change this, also change allocmcache.
		if rate > 0x3fffffff { // make 2*rate not overflow
			rate = 0x3fffffff
		}
		next := int32(fastrand1()) % (2 * int32(rate))
		// Subtract the "remainder" of the current allocation.
		// Otherwise objects that are close in size to sampling rate
		// will be under-sampled, because we consistently discard this remainder.
		next -= (int32(size) - c.next_sample)
		if next < 0 {
			next = 0
		}
		c.next_sample = next
	}

	mProf_Malloc(x, size)
}

type persistentAlloc struct {
	base unsafe.Pointer
	off  uintptr
}

var globalAlloc struct {
	mutex
	persistentAlloc
}

// Wrapper around sysAlloc that can allocate small chunks.
// There is no associated free operation.
// Intended for things like function/type/debug-related persistent data.
// If align is 0, uses default align (currently 8).
func persistentalloc(size, align uintptr, sysStat *uint64) unsafe.Pointer {
	var p unsafe.Pointer
	systemstack(func() {
		p = persistentalloc1(size, align, sysStat)
	})
	return p
}

// Must run on system stack because stack growth can (re)invoke it.
// See issue 9174.
//go:systemstack
func persistentalloc1(size, align uintptr, sysStat *uint64) unsafe.Pointer {
	const (
		chunk    = 256 << 10
		maxBlock = 64 << 10 // VM reservation granularity is 64K on windows
	)

	if size == 0 {
		throw("persistentalloc: size == 0")
	}
	if align != 0 {
		if align&(align-1) != 0 {
			throw("persistentalloc: align is not a power of 2")
		}
		if align > _PageSize {
			throw("persistentalloc: align is too large")
		}
	} else {
		align = 8
	}

	if size >= maxBlock {
		return sysAlloc(size, sysStat)
	}

	mp := acquirem()
	var persistent *persistentAlloc
	if mp != nil && mp.p != 0 {
		persistent = &mp.p.ptr().palloc
	} else {
		lock(&globalAlloc.mutex)
		persistent = &globalAlloc.persistentAlloc
	}
	persistent.off = round(persistent.off, align)
	if persistent.off+size > chunk || persistent.base == nil {
		persistent.base = sysAlloc(chunk, &memstats.other_sys)
		if persistent.base == nil {
			if persistent == &globalAlloc.persistentAlloc {
				unlock(&globalAlloc.mutex)
			}
			throw("runtime: cannot allocate memory")
		}
		persistent.off = 0
	}
	p := add(persistent.base, persistent.off)
	persistent.off += size
	releasem(mp)
	if persistent == &globalAlloc.persistentAlloc {
		unlock(&globalAlloc.mutex)
	}

	if sysStat != &memstats.other_sys {
		mSysStatInc(sysStat, size)
		mSysStatDec(&memstats.other_sys, size)
	}
	return p
}
EOF


# zero garbage stack - kevent in go uses int32 and in c uses int64
cat > runtime/sys_rumprun_amd64.s <<EOF
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// System calls and other sys.stuff for AMD64, NetBSD
// /usr/src/sys/kern/syscalls.master for syscall numbers.
//

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"

// XXX: support for Rumprun is not complete.  If a routine starts
// with SYSCALL, it's a sign that it hasn't been converted.

// We do the backing syscalls two ways:
// 1) directly calling the syscall handler
// 2) indirectly via calling the libc-level syscall wrapper
//
// For historical reasons, some calls use "2".  The new way
// is to do "1" with the eventual target of not having to bring
// the libc layer along for the ride.

// int32 lwp_create(void *context, uintptr flags, void *lwpid)
// XXXTODO
TEXT runtime·lwp_create(SB),NOSPLIT,\$0
	SYSCALL				// XXX: instant trap

	MOVQ	ctxt+0(FP), DI
	MOVQ	flags+8(FP), SI
	MOVQ	lwpid+16(FP), DX
	MOVL	\$309, AX		// sys__lwp_create
	SYSCALL
	JCC	2(PC)
	NEGQ	AX
	MOVL	AX, ret+24(FP)
	RET

// XXXTODO
TEXT runtime·lwp_tramp(SB),NOSPLIT,\$0
	SYSCALL				// XXX: instant trap
	
	// Set FS to point at m->tls.
	LEAQ	m_tls(R8), DI
	CALL	runtime·settls(SB)

	// Set up new stack.
	get_tls(CX)
	MOVQ	R8, g_m(R9)
	MOVQ	R9, g(CX)
	CALL	runtime·stackcheck(SB)

	// Call fn
	CALL	R12

	// It shouldn't return.  If it does, exit.
	MOVL	\$310, AX		// sys__lwp_exit
	SYSCALL
	JMP	-3(PC)			// keep exiting

// XXXTODO
TEXT runtime·osyield(SB),NOSPLIT,\$0
	SYSCALL				// XXX: instant trap

	MOVL	\$350, AX		// sys_sched_yield
	SYSCALL
	RET

TEXT runtime·lwp_park(SB),NOSPLIT,\$0
	MOVQ	\$0, DI				// arg 1 - clock id (REALTIME)
	MOVQ	\$1, SI				// arg 2 - flags (ABSTIME)
	MOVQ	abstime+0(FP), DX		// arg 3 - abstime
	MOVL	unpark+8(FP), CX		// arg 4 - unpark
	MOVQ	hint+16(FP), R8			// arg 5 - hint
	MOVQ	unparkhint+24(FP), R9		// arg 6 - unparkhint
	LEAQ	___lwp_park60(SB), AX
	CALL	AX
	MOVL	AX, ret+32(FP)
	RET

TEXT runtime·lwp_unpark(SB),NOSPLIT,\$0
	MOVL	lwp+0(FP), DI		// arg 1 - lwp
	MOVQ	hint+8(FP), SI		// arg 2 - hint
	LEAQ	_lwp_unpark(SB), AX
	CALL	AX
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime·lwp_self(SB),NOSPLIT,\$0
	LEAQ	_lwp_self(SB), AX
	CALL	AX
	MOVL	AX, ret+0(FP)
	RET

// Exit the entire program (like C exit)
TEXT runtime·exit(SB),NOSPLIT,\$-8
	MOVL	code+0(FP), DI		// arg 1 - exit status
	LEAQ	_exit(SB), AX
	CALL	AX
	MOVL	\$0xf1, 0xf1		// crash
	RET

TEXT runtime·open(SB),NOSPLIT,\$-8
	MOVQ	name+0(FP), DI		// arg 1 pathname
	MOVL	mode+8(FP), SI		// arg 2 flags
	MOVL	perm+12(FP), DX		// arg 3 mode
	LEAQ	_sys_open(SB), AX
	CALL	AX
	MOVL	AX, ret+16(FP)
	RET

TEXT runtime·closefd(SB),NOSPLIT,\$-8
	MOVL	fd+0(FP), DI		// arg 1 fd
	LEAQ	_sys_close(SB), AX
	CALL	AX
	MOVL	AX, ret+8(FP)
	RET

TEXT runtime·read(SB),NOSPLIT,\$-8
	MOVL	fd+0(FP), DI		// arg 1 fd
	MOVQ	p+8(FP), SI		// arg 2 buf
	MOVL	n+16(FP), DX		// arg 3 count
	LEAQ	_sys_read(SB), AX
	CALL	AX
	MOVL	AX, ret+24(FP)
	RET

TEXT runtime·write(SB),NOSPLIT,\$-8
	MOVQ	fd+0(FP), DI		// arg 1 - fd
	MOVQ	p+8(FP), SI		// arg 2 - buf
	MOVL	n+16(FP), DX		// arg 3 - nbyte
	MOVL	\$4, AX			// sys_write
	LEAQ	_sys_write(SB), AX
	CALL	AX
	MOVL	AX, ret+24(FP)
	RET

TEXT runtime·usleep(SB),NOSPLIT,\$16
	MOVL	\$0, DX
	MOVL	usec+0(FP), AX
	MOVL	\$1000000, CX
	DIVL	CX
	MOVQ	AX, 0(SP)		// tv_sec
	MOVL	\$1000, AX
	MULL	DX
	MOVQ	AX, 8(SP)		// tv_nsec

	MOVQ	SP, DI			// arg 1 - rqtp
	MOVQ	\$0, SI			// arg 2 - rmtp
	LEAQ	_sys___nanosleep50(SB), AX
	CALL	AX
	RET

TEXT runtime·raise(SB),NOSPLIT,\$16
	RET

TEXT runtime·raiseproc(SB),NOSPLIT,\$16
	RET

// XXXTODO (can't fix?)
TEXT runtime·setitimer(SB),NOSPLIT,\$-8
	SYSCALL				// XXX: instant trap

	MOVL	mode+0(FP), DI		// arg 1 - which
	MOVQ	new+8(FP), SI		// arg 2 - itv
	MOVQ	old+16(FP), DX		// arg 3 - oitv
	MOVL	\$425, AX		// sys_setitimer
	SYSCALL
	RET

// func now() (sec int64, nsec int32)
TEXT time·now(SB), NOSPLIT, \$32
	MOVQ	\$0, DI			// arg 1 - clock_id
	LEAQ	8(SP), SI		// arg 2 - tp
	LEAQ	_sys___clock_gettime50(SB), AX
	CALL	AX
	MOVQ	8(SP), AX		// sec
	MOVL	16(SP), DX		// nsec

	// sec is in AX, nsec in DX
	MOVQ	AX, sec+0(FP)
	MOVL	DX, nsec+8(FP)
	RET

TEXT runtime·nanotime(SB),NOSPLIT,\$32
	MOVQ	\$0, DI			// arg 1 - clock_id
	LEAQ	8(SP), SI		// arg 2 - tp
	LEAQ	_sys___clock_gettime50(SB), AX
	CALL	AX
	MOVQ	8(SP), AX		// sec
	MOVL	16(SP), DX		// nsec

	// sec is in AX, nsec in DX
	// return nsec in AX
	IMULQ	\$1000000000, AX
	ADDQ	DX, AX
	MOVQ	AX, ret+0(FP)
	RET

// XXXTODO
TEXT runtime·getcontext(SB),NOSPLIT,\$-8
	SYSCALL				// XXX: instant trap

	MOVQ	ctxt+0(FP), DI		// arg 1 - context
	MOVL	\$307, AX		// sys_getcontext
	SYSCALL
	JCC	2(PC)
	MOVL	\$0xf1, 0xf1		// crash
	RET

TEXT runtime·sigprocmask(SB),NOSPLIT,\$0
	RET

TEXT runtime·sigreturn_tramp(SB),NOSPLIT,\$-8
	RET

TEXT runtime·sigaction(SB),NOSPLIT,\$-8
	RET

TEXT runtime·sigtramp(SB),NOSPLIT,\$64
	RET

// for mmap we need to rearrange the args, they are packed in the
// caller's stack frame in a fashion unsuitable for us
TEXT runtime·mmap(SB),NOSPLIT,\$72	// args+retval = 56+16 = 72
	MOVQ	\$197, DI		// mmap
	MOVQ	SP, SI			// args
	MOVQ	addr+0(FP), DX		// .....
	MOVQ	DX, 0(SP)
	MOVQ	len+8(FP), DX
	MOVQ	DX, 8(SP)
	MOVL	prot+16(FP), DX
	MOVQ	DX, 16(SP)
	MOVL	flags+20(FP), DX
	MOVQ	DX, 24(SP)
	MOVL	fd+24(FP), DX
	MOVQ	DX, 32(SP)
	XORQ	DX, DX			// pad for 64bit arg
	MOVQ	DX, 40(SP)
	MOVL	off+28(FP), DX		// offset from Go is 32bit???
	MOVQ	DX, 48(SP)
	MOVQ	\$0, DX			// dlen -- ignored
	MOVQ	SP, CX			// retval[2]
	ADDQ	\$56, CX			// retval[2]
	LEAQ	rump_syscall(SB), AX
	CALL	AX
	TESTQ	AX, AX
	JNE	err
	MOVQ	56(SP), AX		// errno is ignored, clobber with rv
	MOVQ	AX, ret+32(FP)
	RET
 err:
	MOVQ	\$-1, ret+32(FP)
	RET

TEXT runtime·munmap(SB),NOSPLIT,\$16
	MOVQ	\$73, DI
	MOVQ	SP, SI
	ADDQ	\$0x18, SI
	MOVQ	\$0, DX
	MOVQ	SP, CX
	LEAQ	rump_syscall(SB), AX
	CALL	AX
	TESTQ	AX, AX
	JE	2(PC)
	MOVL	\$0xf1, 0xf1		// crash
	RET

// rumprun does not have mad-vice
TEXT runtime·madvise(SB),NOSPLIT,\$0
	RET

TEXT runtime·sigaltstack(SB),NOSPLIT,\$-8
	RET

// set tls base to DI
// XXXTODO
TEXT runtime·settls(SB),NOSPLIT,\$8
	SYSCALL				// XXX: instant trap

	// adjust for ELF: wants to use -8(FS) for g
	ADDQ	\$8, DI			// arg 1 - ptr
	MOVQ	\$317, AX		// sys__lwp_setprivate
	SYSCALL
	JCC	2(PC)
	MOVL	\$0xf1, 0xf1		// crash
	RET

TEXT runtime·sysctl(SB),NOSPLIT,\$16
	MOVQ	\$202, DI
	MOVQ	SP, SI
	ADDQ	\$0x18, SI
	MOVQ	\$0, DX
	MOVQ	SP, CX
	LEAQ	rump_syscall(SB), AX
	CALL	AX
	TESTQ	AX, AX
	JNE	err
	MOVL	0(SP), AX
	MOVL	AX, ret+48(FP)
	RET
 err:
	NEGQ	AX
	MOVL	AX, ret+48(FP)
	RET

// int32 runtime·kqueue(void)
TEXT runtime·kqueue(SB),NOSPLIT,\$16
	MOVQ	\$344, DI
	MOVQ	\$0, SI
	MOVQ	\$0, DX
	MOVQ	SP, CX
	LEAQ	rump_syscall(SB), AX
	CALL	AX
	TESTQ	AX, AX
	JNE	err
	MOVL	0(SP), AX
	MOVL	AX, ret+0(FP)
	RET
 err:
	NEGQ	AX
	MOVL	AX, ret+0(FP)
	RET

// int32 runtime·kevent(int kq, Kevent *changelist, int nchanges, Kevent *eventlist, int nevents, Timespec *timeout)
TEXT runtime·kevent(SB),NOSPLIT,\$16
	MOVQ	\$435, DI		// kevent
	MOVQ	SP, SI			// args
	ADDQ	\$0x18, SI		// args
	MOVL	\$0, 0x04(SI)
	MOVL	\$0, 0x14(SI)
	MOVL	\$0, 0x24(SI)
	MOVQ	\$0, DX			// dlen -- ignored
	MOVQ	SP, CX			// retval
	LEAQ	rump_syscall(SB), AX
	CALL	AX
	TESTQ	AX, AX
	JNE	err
	MOVL	0(SP), AX
	MOVL	AX, ret+48(FP)
	RET
 err:
	NEGQ	AX
	MOVL	AX, ret+48(FP)
	RET

// void runtime·closeonexec(int32 fd)
TEXT runtime·closeonexec(SB),NOSPLIT,\$0
	MOVL	fd+0(FP), DI	// fd
	MOVQ	\$2, SI		// F_SETFD
	MOVQ	\$1, DX		// FD_CLOEXEC
	LEAQ	_sys_fcntl(SB), AX
	CALL	AX
	RET

EOF