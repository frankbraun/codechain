// NOTE: cgo can't generate struct Stat_t and struct Statfs_t yet
// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs types_darwin.go

// +build arm,darwin

package unix

type Termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]uint8
	Ispeed uint32
	Ospeed uint32
}
