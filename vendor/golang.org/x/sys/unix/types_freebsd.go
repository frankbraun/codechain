// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

/*
Input to cgo -godefs.  See README.md
*/

// +godefs map struct_in_addr [4]byte /* in_addr */
// +godefs map struct_in6_addr [16]byte /* in6_addr */

package unix

/*
#define KERNEL
#include <dirent.h>
#include <fcntl.h>
#include <poll.h>
#include <signal.h>
#include <termios.h>
#include <stdio.h>
#include <unistd.h>
#include <sys/capability.h>
#include <sys/event.h>
#include <sys/mman.h>
#include <sys/mount.h>
#include <sys/param.h>
#include <sys/ptrace.h>
#include <sys/resource.h>
#include <sys/select.h>
#include <sys/signal.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/un.h>
#include <sys/utsname.h>
#include <sys/wait.h>
#include <net/bpf.h>
#include <net/if.h>
#include <net/if_dl.h>
#include <net/route.h>
#include <netinet/in.h>
#include <netinet/icmp6.h>
#include <netinet/tcp.h>

enum {
	sizeofPtr = sizeof(void*),
};

union sockaddr_all {
	struct sockaddr s1;	// this one gets used for fields
	struct sockaddr_in s2;	// these pad it out
	struct sockaddr_in6 s3;
	struct sockaddr_un s4;
	struct sockaddr_dl s5;
};

struct sockaddr_any {
	struct sockaddr addr;
	char pad[sizeof(union sockaddr_all) - sizeof(struct sockaddr)];
};

// This structure is a duplicate of stat on FreeBSD 8-STABLE.
// See /usr/include/sys/stat.h.
struct stat8 {
#undef st_atimespec	st_atim
#undef st_mtimespec	st_mtim
#undef st_ctimespec	st_ctim
#undef st_birthtimespec	st_birthtim
	__dev_t   st_dev;
	ino_t     st_ino;
	mode_t    st_mode;
	nlink_t   st_nlink;
	uid_t     st_uid;
	gid_t     st_gid;
	__dev_t   st_rdev;
#if __BSD_VISIBLE
	struct  timespec st_atimespec;
	struct  timespec st_mtimespec;
	struct  timespec st_ctimespec;
#else
	time_t    st_atime;
	long      __st_atimensec;
	time_t    st_mtime;
	long      __st_mtimensec;
	time_t    st_ctime;
	long      __st_ctimensec;
#endif
	off_t     st_size;
	blkcnt_t st_blocks;
	blksize_t st_blksize;
	fflags_t  st_flags;
	__uint32_t st_gen;
	__int32_t st_lspare;
#if __BSD_VISIBLE
	struct timespec st_birthtimespec;
	unsigned int :(8 / 2) * (16 - (int)sizeof(struct timespec));
	unsigned int :(8 / 2) * (16 - (int)sizeof(struct timespec));
#else
	time_t    st_birthtime;
	long      st_birthtimensec;
	unsigned int :(8 / 2) * (16 - (int)sizeof(struct __timespec));
	unsigned int :(8 / 2) * (16 - (int)sizeof(struct __timespec));
#endif
};

// This structure is a duplicate of if_data on FreeBSD 8-STABLE.
// See /usr/include/net/if.h.
struct if_data8 {
	u_char  ifi_type;
	u_char  ifi_physical;
	u_char  ifi_addrlen;
	u_char  ifi_hdrlen;
	u_char  ifi_link_state;
	u_char  ifi_spare_char1;
	u_char  ifi_spare_char2;
	u_char  ifi_datalen;
	u_long  ifi_mtu;
	u_long  ifi_metric;
	u_long  ifi_baudrate;
	u_long  ifi_ipackets;
	u_long  ifi_ierrors;
	u_long  ifi_opackets;
	u_long  ifi_oerrors;
	u_long  ifi_collisions;
	u_long  ifi_ibytes;
	u_long  ifi_obytes;
	u_long  ifi_imcasts;
	u_long  ifi_omcasts;
	u_long  ifi_iqdrops;
	u_long  ifi_noproto;
	u_long  ifi_hwassist;
// FIXME: these are now unions, so maybe need to change definitions?
#undef ifi_epoch
	time_t  ifi_epoch;
#undef ifi_lastchange
	struct  timeval ifi_lastchange;
};

// This structure is a duplicate of if_msghdr on FreeBSD 8-STABLE.
// See /usr/include/net/if.h.
struct if_msghdr8 {
	u_short ifm_msglen;
	u_char  ifm_version;
	u_char  ifm_type;
	int     ifm_addrs;
	int     ifm_flags;
	u_short ifm_index;
	struct  if_data8 ifm_data;
};
*/
import "C"

// Terminal handling

type Termios C.struct_termios
