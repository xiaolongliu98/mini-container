package cgroup

import (
	"encoding/json"
	"fmt"
	"mini-container/common"
	"mini-container/config"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type CgroupType string

const (
	CgroupCpu CgroupType = "cpu"
	CgroupMem CgroupType = "memory"
)

type ICgroup interface {
	Type() CgroupType
	ContainerName() string
	Apply(childPID int) error
	json.Marshaler
	json.Unmarshaler
}

func Apply(cg ICgroup, childPID int) error {
	applied, err := Applied(cg, childPID)
	if err != nil {
		return err
	}
	if applied {
		return nil
	}
	if err = Release(cg); err != nil {
		return err
	}

	return cg.Apply(childPID)
}

func Applied(cg ICgroup, childPID int) (bool, error) {
	path := filepath.Join(CgroupPath(cg), string(cg.Type()), "tasks")
	if !common.IsExistPath(path) {
		return false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return string(data) == strconv.Itoa(childPID), nil
}

func Release(cg ICgroup) error {
	return clearCgroup(cg.ContainerName(), cg.Type())
}

// CgroupPath format: /sys/fs/cgroup/[type]/[projName]/[containerName]
func CgroupPath(cg ICgroup) string {
	return filepath.Join(config.CgroupsDir, string(cg.Type()), config.ProjName, cg.ContainerName())
}

func createCgroup(name string) error {
	cgroupPath := filepath.Join(config.CgroupsDir, name)
	err := os.MkdirAll(cgroupPath, 0755)
	return err
}

// TODO 可能需要处理未找到的错误情况，这种情况视为无错误
func clearCgroup(name string, cgroupType CgroupType) error {
	output, err := exec.Command("cgdelete", "-r", fmt.Sprintf("%s:%s/%s", cgroupType, config.ProjName, name)).Output()
	if err != nil {
		return fmt.Errorf("clear %s cgroup fail 1 err=%s output=%s", cgroupType, err, string(output))
	}
	return os.RemoveAll(filepath.Join(config.CgroupsDir, string(cgroupType), config.ProjName, name))
}
