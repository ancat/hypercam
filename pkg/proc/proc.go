package proc

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"encoding/hex"
	"io/fs"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var SYS_EXECVEAT, EXECVEAT_ERR = execveat_syscall_number()
func SneakyExec(handle *os.File, argv []string) {
	if EXECVEAT_ERR != nil {
		panic(EXECVEAT_ERR)
	}

	empty_path, _ := syscall.BytePtrFromString("")
	p_empty_path := unsafe.Pointer(empty_path)
	p_argv, _ := syscall.SlicePtrFromStrings(argv)
	syscall.Syscall6(
		uintptr(SYS_EXECVEAT),
		uintptr(handle.Fd()),
		uintptr(p_empty_path), // if empty, we execute the fd instead
		uintptr(unsafe.Pointer(&p_argv[0])),
		uintptr(0),
		uintptr(0x1000), // AT_EMPTY_PATH
		uintptr(0),
	)
}

func execveat_syscall_number() (int, error) {
	handle, err := os.Open("/usr/include/asm/unistd_64.h")
	if err != nil {
		return -1, err
	}

	scanner := bufio.NewScanner(handle)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[1] == "__NR_execveat" {
			return strconv.Atoi(fields[2])
		}
	}

	return -1, errors.New("couldn't find it I guess")
}

func GetExe(pid int) string {
	fd_name := fmt.Sprintf("/proc/%d/exe", pid)
	link, err := os.Readlink(fd_name)

	if err != nil {
		panic(err)
	}

	return link
}

func GetComm(pid int) string {
	path := fmt.Sprintf("/proc/%d/cmdline", pid)
	comm, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(bytes.TrimSpace(comm))
}

func GetFds(pid int) ([]string, []Connection) {
	handle, _ := os.Open(fmt.Sprintf("/proc/%d/fd", pid))
	names, _ := handle.Readdirnames(0)
	file_list := make([]string, 0)
	socket_list := make([]Connection, 0)
	tcp_table := GetTcp4Table(pid)

	for _, name := range(names) {
		fd_name := fmt.Sprintf("/proc/%d/fd/%s", pid, name)
		link, _ := os.Readlink(fd_name)
		stat, _ := os.Stat(fd_name)
		if stat.Mode()&fs.ModeSocket != 0 {
			var socket_inode int
			n, err := fmt.Sscanf(link, "socket:[%d]", &socket_inode)
			if n == 0 {
				panic(err)
			}

			socket_list = append(socket_list, tcp_table[socket_inode])
		}

		if stat.Mode().IsRegular() && !strings.HasPrefix(link, "anon_inode") {
			file_list = append(file_list, link)
		}
	}

	return file_list, socket_list
}

func decode_ipv4_address(ip string) string {
	fields := strings.Split(ip, ":")
	integer_ip, _ := hex.DecodeString(fields[0])
	integer_port, _ := strconv.ParseUint(fields[1], 16, 16)

	decoded_ip := net.IPv4(integer_ip[3], integer_ip[2], integer_ip[1], integer_ip[0])
	return fmt.Sprintf("%s:%d", decoded_ip.String(), integer_port)
}

func GetTcp4Table(pid int) map[int]Connection {
	inode_table := make(map[int]Connection)
	handle, err := os.Open(fmt.Sprintf("/proc/%d/net/tcp", pid))
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(handle)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if fields[1] == "local_address" {
			continue
		}

		local_address := decode_ipv4_address(fields[1])
		remote_address := decode_ipv4_address(fields[2])
		inode, _ := strconv.Atoi(fields[9])
		inode_table[inode] = Connection{local_address, remote_address};
	}

	return inode_table
}

func GetPidRoot(pid int) (*os.File, error) {
	handle, err := os.Open(fmt.Sprintf("/proc/%d/root", pid))
	if err != nil {
		return nil, err
	}

	return handle, nil
}
