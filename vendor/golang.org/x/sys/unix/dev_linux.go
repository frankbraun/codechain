// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Functions to access/create device major and minor numbers matching the
// encoding used by the Linux kernel and glibc.
//
// The information below is extracted and adapted from bits/sysmacros.h in the
// glibc sources:
//
// dev_t in glibc is 64-bit, with 32-bit major and minor numbers. glibc's
// default encoding is MMMM Mmmm mmmM MMmm, where M is a hex digit of the major
// number and m is a hex digit of the minor number. This is backward compatible
// with legacy systems where dev_t is 16 bits wide, encoded as MMmm. It is also
// backward compatible with the Linux kernel, which for some architectures uses
// 32-bit dev_t, encoded as mmmM MMmm.

package unix
