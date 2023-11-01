package cgroup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

type memoryCgroupAlias struct {
	ContainerName string
	Limit         uint
}

type MemoryCgroup struct {
	containerName string
	limit         uint // MiB
}

func (cg *MemoryCgroup) MarshalJSON() ([]byte, error) {
	return json.Marshal(memoryCgroupAlias{
		ContainerName: cg.containerName,
		Limit:         cg.limit,
	})
}

func (cg *MemoryCgroup) UnmarshalJSON(bytes []byte) error {
	var alias memoryCgroupAlias
	err := json.Unmarshal(bytes, &alias)
	if err != nil {
		return err
	}
	cg.containerName = alias.ContainerName
	cg.limit = alias.Limit
	return nil
}

func NewMemoryCgroup(containerName string, limitMiB uint) *MemoryCgroup {
	return &MemoryCgroup{
		containerName: containerName,
		limit:         limitMiB,
	}
}

func (cg *MemoryCgroup) Type() CgroupType {
	return CgroupMem
}

func (cg *MemoryCgroup) ContainerName() string {
	return cg.containerName
}

func (cg *MemoryCgroup) Apply(childPID int) error {
	path := CgroupPath(cg)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	// write task
	err = os.WriteFile(filepath.Join(path, "tasks"), []byte(strconv.Itoa(childPID)), 0644)
	if err != nil {
		return err
	}
	// write memory.limit_in_bytes
	err = os.WriteFile(filepath.Join(path, "memory.limit_in_bytes"), []byte(strconv.Itoa(int(cg.limit*1024*1024))), 0700)
	if err != nil {
		return err
	}
	return nil
}
