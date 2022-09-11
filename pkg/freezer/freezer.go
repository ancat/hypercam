package freezer

import (
    "bufio"
	"bytes"
    "fmt"
    //"io/fs"
    //"errors"
    "strconv"
    "strings"
	"os"
)

/*

cat /sys/fs/cgroup/pids/docker/902e76b310599cce945cdcb6810d02dd8ed2b45ff7d2f833bbcecea2b78d84f4/cgroup.procs

*/

func GetFreezerInfo(pid int) (string, error) {
    handle, err := os.Open(fmt.Sprintf("/proc/%d/cgroup", pid))
    if err != nil {
        return "", err
    }

    defer handle.Close()
    scanner := bufio.NewScanner(handle)

    for scanner.Scan() {
        s := strings.SplitN(scanner.Text(), ":", 3)
        if s[1] == "freezer" {
            return s[2], nil
        }
    }

    return "", nil
}

func GetPidsByCgroup(name string) []int {
    path := fmt.Sprintf("/sys/fs/cgroup/pids/%s/tasks", name);
    tasks, err := os.ReadFile(path)
    if err != nil {
        panic(err)
    }

    var pids []int
    for _, line := range strings.Split(string(tasks), "\n") {
        if line == "" {
            continue
        }

        pid, err := strconv.Atoi(line)
        if err != nil || pid == 0 {
            continue
        }

        pids = append(pids, pid)
    }

    return pids
}

func UpdateFreezerTasks(name string, pids []int) {
    err := os.MkdirAll(fmt.Sprintf("/sys/fs/cgroup/freezer/%s", name), 0755)
    if err != nil {
        panic(err)
    }

    handle, err := os.OpenFile(fmt.Sprintf("/sys/fs/cgroup/freezer/%s/tasks", name), os.O_RDWR|os.O_CREATE, 0755)
    if err != nil {
        panic(err)
    }
    defer handle.Close()

    for _, pid := range pids {
        handle.WriteString(fmt.Sprintf("%d\n", pid))
    }
}

func GetFreezerStateByName(name string) string {
    state, err := os.ReadFile(fmt.Sprintf("/sys/fs/cgroup/freezer/%s/freezer.state", name))
    if err != nil {
        panic(err)
    }

    return string(bytes.TrimSpace(state))
}

func GetFreezerStateByPid(pid int) string {
    cgroup_name, err := GetFreezerInfo(pid)
    if err != nil {
        panic(err)
    }

    return GetFreezerStateByName(cgroup_name)
}

func UpdateFreezerStateByName(name string, state string) {
    handle, err := os.OpenFile(fmt.Sprintf("/sys/fs/cgroup/freezer/%s/freezer.state", name), os.O_RDWR, 0755)
    if err != nil {
        panic(err)
    }
    defer handle.Close()

    handle.WriteString(fmt.Sprintf("%s\n", state))
}

/*
 41 def update_freezer_state_by_pid(pid, state):
 42     return update_freezer_state_by_cgroup(get_cgroup_info(pid), state)
*/
func UpdateFreezerStateByPid(pid int, state string) {}

