// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix

// Aix system calls.
// This file is compiled as ordinary Go code,
// but it is also input to mksyscall,
// which parses the //sys lines and generates system call stubs.
// Note that sometimes we use a lowercase //sys name and
// wrap it in our own nicer implementation.

package unix

import (
	"unsafe"
)

func ioctlSetTermios(fd int, req uint, value *Termios) error {
	return ioctl(fd, req, uintptr(unsafe.Pointer(value)))
}

func IoctlGetTermios(fd int, req uint) (*Termios, error) {
	var value Termios
	err := ioctl(fd, req, uintptr(unsafe.Pointer(&value)))
	return &value, err
}
