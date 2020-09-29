// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build arm

package atomic

import (
	"internal/cpu"
	"unsafe"
)

// Export some functions via linkname to assembly in sync/atomic.
//go:linkname Xchg
//go:linkname Xchguintptr

type spinlock struct {
	v uint32
}

//go:nosplit
func (l *spinlock) lock() {
	for {
		if Cas(&l.v, 0, 1) {
			return
		}
	}
}

//go:nosplit
func (l *spinlock) unlock() {
	Store(&l.v, 0)
}

var locktab [57]struct {
	l   spinlock
	pad [cpu.CacheLinePadSize - unsafe.Sizeof(spinlock{})]byte
}

func addrLock(addr *uint64) *spinlock {
	return &locktab[(uintptr(unsafe.Pointer(addr))>>3)%uintptr(len(locktab))].l
}

// Atomic add and return new value.
//go:nosplit
func Xadd(val *uint32, delta int32) uint32 {
	for {
		oval := *val
		nval := oval + uint32(delta)
		if Cas(val, oval, nval) {
			return nval
		}
	}
}

//go:noescape
func Xadduintptr(ptr *uintptr, delta uintptr) uintptr

//go:nosplit
func Xchg(addr *uint32, v uint32) uint32 {
	for {
		old := *addr
		if Cas(addr, old, v) {
			return old
		}
	}
}

//go:nosplit
func Xchguintptr(addr *uintptr, v uintptr) uintptr {
	return uintptr(Xchg((*uint32)(unsafe.Pointer(addr)), uint32(v)))
}

// Not noescape -- it installs a pointer to addr.
func StorepNoWB(addr unsafe.Pointer, v unsafe.Pointer)

//go:noescape
func Store(addr *uint32, v uint32)

//go:noescape
func StoreRel(addr *uint32, v uint32)

//go:nosplit
//Go中的CAS方法
//比较并交换(compare and swap, CAS)，是原子操作的一种，可用于在多线程编程中实现不被打断的数据交换操作，从而避免多线程同时改写某一数据时由于执行顺序不确定性以及中断的不可预知性产生的数据不一致问题。 该操作通过将内存中的值与指定数据进行比较，当数值一样时将内存中的数据替换为新的值。
//获取的乐观锁
func goCas64(addr *uint64, old, new uint64) bool {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var ok bool
	addrLock(addr).lock()
	if *addr == old {
		*addr = new
		ok = true
	}
	addrLock(addr).unlock()
	return ok
}

//go:nosplit
//加法的乐观锁
func goXadd64(addr *uint64, delta int64) uint64 {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var r uint64
	addrLock(addr).lock()
	r = *addr + uint64(delta)
	*addr = r
	addrLock(addr).unlock()
	return r
}

//go:nosplit
//乘法的乐观锁
func goXchg64(addr *uint64, v uint64) uint64 {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var r uint64
	addrLock(addr).lock()
	r = *addr
	*addr = v
	addrLock(addr).unlock()
	return r
}

//go:nosplit
//赋值的乐观锁
func goLoad64(addr *uint64) uint64 {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	var r uint64
	addrLock(addr).lock()
	r = *addr
	addrLock(addr).unlock()
	return r
}

//go:nosplit
//存储的乐观锁
func goStore64(addr *uint64, v uint64) {
	if uintptr(unsafe.Pointer(addr))&7 != 0 {
		*(*int)(nil) = 0 // crash on unaligned uint64
	}
	_ = *addr // if nil, fault before taking the lock
	addrLock(addr).lock()
	*addr = v
	addrLock(addr).unlock()
}

//go:nosplit
//乐观锁下的或
func Or8(addr *uint8, v uint8) {
	// Align down to 4 bytes and use 32-bit CAS.
	uaddr := uintptr(unsafe.Pointer(addr))
	addr32 := (*uint32)(unsafe.Pointer(uaddr &^ 3))
	word := uint32(v) << ((uaddr & 3) * 8) // little endian
	for {
		old := *addr32
		if Cas(addr32, old, old|word) {
			return
		}
	}
}

//go:nosplit
//乐观锁下的与
func And8(addr *uint8, v uint8) {
	// Align down to 4 bytes and use 32-bit CAS.
	uaddr := uintptr(unsafe.Pointer(addr))
	addr32 := (*uint32)(unsafe.Pointer(uaddr &^ 3))
	word := uint32(v) << ((uaddr & 3) * 8)    // little endian
	mask := uint32(0xFF) << ((uaddr & 3) * 8) // little endian
	word |= ^mask
	for {
		old := *addr32
		if Cas(addr32, old, old&word) {
			return
		}
	}
}

//go:nosplit
func armcas(ptr *uint32, old, new uint32) bool

//go:noescape
func Load(addr *uint32) uint32

// NO go:noescape annotation; *addr escapes if result escapes (#31525)
func Loadp(addr unsafe.Pointer) unsafe.Pointer

//go:noescape
func Load8(addr *uint8) uint8

//go:noescape
func LoadAcq(addr *uint32) uint32

//go:noescape
func Cas64(addr *uint64, old, new uint64) bool

//go:noescape
func CasRel(addr *uint32, old, new uint32) bool

//go:noescape
func Xadd64(addr *uint64, delta int64) uint64

//go:noescape
func Xchg64(addr *uint64, v uint64) uint64

//go:noescape
func Load64(addr *uint64) uint64

//go:noescape
func Store8(addr *uint8, v uint8)

//go:noescape
func Store64(addr *uint64, v uint64)
