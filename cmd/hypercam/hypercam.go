package hypercam

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unicode"

	"github.com/ancat/hypercam/pkg/freezer"
	"github.com/ancat/hypercam/pkg/proc"
	"github.com/ancat/hypercam/pkg/ns"
)

func Pr(pid int) error {
	cgroup_name, err := freezer.GetFreezerInfo(pid)
	if err != nil {
		return errors.New("no such pid")
	}

	var pids []int
	if cgroup_name == "" || cgroup_name == "/" {
		pids = []int{pid}
	} else {
		pids = freezer.GetPidsByCgroup(cgroup_name)
	}

	for _, pid := range pids {
		fmt.Printf("%s (%s) %d\n", proc.GetComm(pid), proc.GetExe(pid), pid)

		files, sockets := proc.GetFds(pid)
		for _, file := range files {
			fmt.Printf("- Open File: %s\n", file)
		}

		for _, socket := range sockets {
			fmt.Printf("- Open Socket: %s<->%s\n", socket.LocalAddress, socket.RemoteAddress)
		}
	}

	return nil
}

func SpawnShellInside(pid int, portal bool, host_executable string, guest_executable string) {
	var handle *os.File
	var err error

	handle, err = ns.GetFdForPidNs(pid, "cgroup")
	ns.SetNs(handle.Fd(), ns.CLONE_NEWCGROUP)
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	handle, err = ns.GetFdForPidNs(pid, "ipc")
	ns.SetNs(handle.Fd(), ns.CLONE_NEWIPC)
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	handle, err = ns.GetFdForPidNs(pid, "uts")
	ns.SetNs(handle.Fd(), ns.CLONE_NEWUTS)
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	handle, err = ns.GetFdForPidNs(pid, "net")
	ns.SetNs(handle.Fd(), ns.CLONE_NEWNET)
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	handle, err = ns.GetFdForPidNs(pid, "pid")
	ns.SetNs(handle.Fd(), ns.CLONE_NEWPID)
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	/*handle, _ = get_fd_for_pid_ns(4106, "mnt")
	ns.SetNs(handle.Fd(), ns.CLONE_NEWNS)

	handle, _ = get_fd_for_pid_ns(4106, "user")
	ns.SetNs(handle.Fd(), ns.CLONE_NEWUSER)*/

	if portal {
		portal_path := fmt.Sprintf("/proc/%d/cwd/portal420", pid)
		host_path, _ := os.MkdirTemp("", "portal")

		os.Mkdir(portal_path, 0700)
		os.RemoveAll(host_path)
		os.Symlink(portal_path, host_path)
		fmt.Printf("Host Endpoint: %s\nGuest Endpoint: %s\n", host_path, "/portal420")
	}

	var sneaky_handle *os.File
	if host_executable != "" {
		sneaky_handle, err = os.Open(host_executable)
		if err != nil {
			panic(err)
		}
	}

	dir, _ := proc.GetPidRoot(4106)
	syscall.Fchdir(int(dir.Fd()))
	syscall.Chroot(".")

	if host_executable != "" {
		proc.SneakyExec(sneaky_handle, []string{"hypercam shell"})
		panic("FAIL")
	}

	if guest_executable == "" {
		guest_executable = "/bin/sh"
	}

	syscall.Exec(guest_executable, nil, nil)
	panic("FAIL")
}

func DumpMaps(pid int, hex_dump bool) {
	maps := proc.GetMaps(pid)
	for _, memory_map := range maps {
		var page *proc.MemoryMap
		if memory_map.Name == "[stack]" {
			page = memory_map
			fmt.Printf("Stack Located\n")
		}

		if memory_map.Name == "[heap]" {
			page = memory_map
			fmt.Printf("Heap Located\n")
		}

		if page == nil {
			continue
		}

		page_contents, err := proc.ReadProcessMemory(
			pid,
			memory_map.Base,
			memory_map.Length,
		)

		if err != nil {
			fmt.Printf("Failed to read page: %s\n", err)
			continue
		}

		if hex_dump {
			fmt.Printf("%s", hex.Dump(page_contents))
		} else {
			print_strings(page_contents, 10)
		}
	}
}

func print_strings(buffer []byte, min_len int) {
	start := 0
	for i, c := range(buffer) {
		if !unicode.IsPrint(rune(c)) {
			strlen := i - start
			if strlen >= min_len {
				trimmed := strings.TrimSpace(string(buffer[start:i]))
				if len(trimmed) >= min_len {
					fmt.Printf("%s\n", trimmed)
				}
			}

			start = i
		}
	}
}
