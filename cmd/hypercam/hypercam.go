package hypercam

import (
	"errors"
	"fmt"
	"os"
	"syscall"

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

func SpawnShellInside(pid int, portal bool) {
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
		fmt.Printf("one portal coming right up boss\n")
		portal_path := fmt.Sprintf("/proc/%d/cwd/portal420", pid)
		host_path, _ := os.MkdirTemp("", "portal")

		os.Mkdir(portal_path, 0700)
		os.RemoveAll(host_path)
		os.Symlink(portal_path, host_path)
		fmt.Printf("portal here: %s\n", host_path)
	} else {
		fmt.Printf("no portals tonight\n")
	}

	dir, _ := proc.GetPidRoot(4106)
	syscall.Fchdir(int(dir.Fd()))
	syscall.Chroot(".")
	syscall.Exec("/bin/bash", nil, nil)

}
