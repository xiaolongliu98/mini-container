package cgroup

import "C"
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

type cpuCgroupAlias struct {
	ContainerName string
	Percent       uint
}

type CPUCgroup struct {
	containerName string
	percent       uint // percent range: [1, 100]
}

func (cg *CPUCgroup) MarshalJSON() ([]byte, error) {
	return json.Marshal(cpuCgroupAlias{
		ContainerName: cg.containerName,
		Percent:       cg.percent,
	})
}

func (cg *CPUCgroup) UnmarshalJSON(bytes []byte) error {
	var alias cpuCgroupAlias
	err := json.Unmarshal(bytes, &alias)
	if err != nil {
		return err
	}
	cg.containerName = alias.ContainerName
	cg.percent = alias.Percent
	return nil
}

func NewCPUCgroup(containerName string, percent uint) *CPUCgroup {
	if percent < 1 || percent > 100 {
		panic("percent range: [1, 100]")
	}

	return &CPUCgroup{
		containerName: containerName,
		percent:       percent,
	}
}

func (cg *CPUCgroup) Type() CgroupType {
	return CgroupCpu
}

func (cg *CPUCgroup) ContainerName() string {
	return cg.containerName
}

func (cg *CPUCgroup) Apply(childPID int) error {
	path := CgroupPath(cg)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	// 设置cpu
	if err := os.WriteFile(filepath.Join(path, "tasks"), []byte(strconv.Itoa(childPID)), 0644); err != nil {
		return err
	}

	basePeriod := 100_000 // 100ms
	quota := int(float64(basePeriod) * float64(cg.percent) / 100.0)

	err = os.WriteFile(filepath.Join(path, "/cpu.cfs_period_us"), []byte(strconv.Itoa(basePeriod)), 0644)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(path, "/cpu.cfs_quota_us"), []byte(strconv.Itoa(quota)), 0644)
	if err != nil {
		return err
	}

	return nil
}
