package cgroup

import (
	"mini-container/common"
	"os"
	"path/filepath"
	"strconv"
)

const (
	Cgroups = "/sys/fs/cgroup/"
)

func CreateCgroup(name string) error {
	cgroupPath := filepath.Join(Cgroups, name)
	err := os.Mkdir(cgroupPath, 0755)
	return err
}

// DeleteCgroup 删除 cgroup, 但是只能删除空的 cgroup
func DeleteCgroup(name string) error {
	cgroupPath := filepath.Join(Cgroups, name)
	err := os.Remove(cgroupPath)
	return err
}

// DeleteCgroupForce 删除 cgroup, 但是可以删除非空的 cgroup
func DeleteCgroupForce(name string) error {
	cgroupPath := filepath.Join(Cgroups, name)
	err := os.RemoveAll(cgroupPath)
	return err
}

func cg() {

	pids := filepath.Join(Cgroups, "pids")
	common.Must(os.Mkdir(filepath.Join(pids, "lizrice"), 0755))
	common.Must(os.WriteFile(filepath.Join(pids, "lizrice/pids.max"), []byte("20"), 0700))
	common.Must(os.WriteFile(filepath.Join(pids, "lizrice/notify_on_release"), []byte("1"), 0700))
	common.Must(os.WriteFile(filepath.Join(pids, "lizrice/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))

	cpu := filepath.Join(Cgroups, "cpu")
	common.Must(os.Mkdir(filepath.Join(cpu, "lizrice"), 0755))

	common.Must(os.WriteFile(filepath.Join(cpu, "lizrice/cpu.cfs_quota_us"), []byte("50000"), 0700))
	common.Must(os.WriteFile(filepath.Join(cpu, "lizrice/notify_on_release"), []byte("1"), 0700))
	common.Must(os.WriteFile(filepath.Join(cpu, "lizrice/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}
