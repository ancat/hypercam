package ns

import (
	"bufio"
	"errors"
    "fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var CLONE_THREAD       = 0x00010000
var CLONE_NEWNS        = 0x00020000
var CLONE_CHILD_SETTID = 0x01000000
var CLONE_NEWCGROUP    = 0x02000000
var CLONE_NEWUTS       = 0x04000000
var CLONE_NEWIPC       = 0x08000000
var CLONE_NEWUSER      = 0x10000000
var CLONE_NEWPID       = 0x20000000
var CLONE_NEWNET       = 0x40000000
var CLONE_IO           = 0x80000000
var SYS_SETNS, NSERR   = setns_syscall_number()

func SetNs(fd uintptr, nstype int) {
	if NSERR != nil {
		panic(NSERR)
	}

	syscall.Syscall(uintptr(SYS_SETNS), uintptr(fd), uintptr(nstype), 0)
}

func GetFdForPidNs(pid int, ns string) (*os.File, error) {
	handle, err := os.Open(fmt.Sprintf("/proc/%d/ns/%s", pid, ns))
	if err != nil {
		return nil, err
	}

	return handle, nil
}

func setns_syscall_number() (int, error) {
	handle, err := os.Open("/usr/include/asm/unistd_64.h")
	if err != nil {
		return -1, err
	}

	scanner := bufio.NewScanner(handle)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[1] == "__NR_setns" {
			return strconv.Atoi(fields[2])
		}
	}

	return -1, errors.New("couldn't find it I guess")
}
