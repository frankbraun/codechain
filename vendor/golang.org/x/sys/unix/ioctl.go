// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package unix

import "runtime"

// IoctlSetTermios performs an ioctl on fd with a *Termios.
//
// The req value will usually be TCSETA or TIOCSETA.
func IoctlSetTermios(fd int, req uint, value *Termios) error {
	// TODO: if we get the chance, remove the req parameter.
	err := ioctlSetTermios(fd, req, value)
	runtime.KeepAlive(value)
	return err
}
