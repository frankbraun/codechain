// +build sparc64,linux
// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs types_linux.go | go run mkpost.go

package unix

type Termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Line   uint8
	Cc     [19]uint8
	Ispeed uint32
	Ospeed uint32
}
