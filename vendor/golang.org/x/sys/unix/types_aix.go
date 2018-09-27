// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore
// +build aix

/*
Input to cgo -godefs.  See also mkerrors.sh and mkall.sh
*/

// +godefs map struct_in_addr [4]byte /* in_addr */
// +godefs map struct_in6_addr [16]byte /* in6_addr */

package unix

/*
#include <sys/types.h>
#include <sys/time.h>
#include <sys/limits.h>
#include <sys/un.h>
#include <utime.h>
#include <sys/utsname.h>
#include <sys/poll.h>
#include <sys/resource.h>
#include <sys/stat.h>
#include <sys/statfs.h>
#include <sys/termio.h>
#include <sys/ioctl.h>

#include <termios.h>

#include <net/if.h>
#include <net/if_dl.h>
#include <netinet/in.h>
#include <netinet/icmp6.h>


#include <dirent.h>
#include <fcntl.h>

enum {
	sizeofPtr = sizeof(void*),
};

union sockaddr_all {
	struct sockaddr s1;     // this one gets used for fields
	struct sockaddr_in s2;  // these pad it out
	struct sockaddr_in6 s3;
	struct sockaddr_un s4;
	struct sockaddr_dl s5;
};

struct sockaddr_any {
	struct sockaddr addr;
	char pad[sizeof(union sockaddr_all) - sizeof(struct sockaddr)];
};

*/
import "C"

// Terminal handling

type Termios C.struct_termios
