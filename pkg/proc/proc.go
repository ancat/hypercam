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

	"golang.org/x/sys/unix"
)

var SYS_EXECVEAT, EXECVEAT_ERR = execveat_syscall_number()
func ReadProcessMemory(pid int, address uintptr, length uint64) ([]byte, error) {
	destination := make([]byte, length)
	// _ = new(unix.Iovec)
	local_iovec := make([]unix.Iovec, 1)
	local_iovec[0].Base = (*byte) (unsafe.Pointer(&destination[0]))
	local_iovec[0].Len = length

	remote_iovec := make([]unix.RemoteIovec, 1)
	remote_iovec[0].Base = address
	remote_iovec[0].Len = int(length) // why is local a uint64 and remote int¿

	_, err := unix.ProcessVMReadv(pid, local_iovec, remote_iovec, 0)
	return destination, err
}

func GetMaps(pid int) []*MemoryMap {
	handle, err := os.Open(fmt.Sprintf("/proc/%d/maps", pid))
	if err != nil {
		panic(err)
	}

	defer handle.Close()
	scanner := bufio.NewScanner(handle)
	maps := make([]*MemoryMap, 0)
	for scanner.Scan() {
		line := scanner.Text()
		cur_map := new(MemoryMap)

		var t_map_start uint64
		var t_map_end uint64
		var t_perms string
		var t_len uint64
		var t_dev string
		var t_inode int
		var t_name string

		// 7ffd4d9f1000-7ffd4d9f3000 r-xp 00000000 00:00 0				[vdso]
		scanned, _ := fmt.Sscanf(line, "%x-%x %s %x %s %d %s",
			&t_map_start,
			&t_map_end,
			&t_perms,
			&t_len,
			&t_dev,
			&t_inode,
			&t_name)

		if scanned != 7 {
			continue
		}

		// address_space := fields[0]
		perms := 0
		for i := 0; i < 3; i++ {
			switch t_perms[i] {
			case 'r':
				perms |= 0x1
			case 'w':
				perms |= 0x2
			case 'x':
				perms |= 0x4
			}
		}

		cur_map.Name = t_name
		cur_map.Base = uintptr(t_map_start)
		cur_map.Length = t_map_end - t_map_start
		cur_map.Perms = perms
		maps = append(maps, cur_map)
	}

	return maps
}

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
