package container

import (
	"mini-container/common"
	"mini-container/config"
	"mini-container/internal/cgroup"
	"mini-container/internal/fs"
	"mini-container/internal/network"
	"os"
	"path/filepath"
)

func ExistsContainer(name string) bool {
	return common.IsExistPath(filepath.Join(config.ContainerConfigDir, name, ConfigName)) ||
		common.IsExistPath(filepath.Join(config.ContainerConfigDir, name, StateName))
}

// NewCreatedContainer 创建一个创建状态的容器
// 注意：调用前需要确保容器不存在
func NewCreatedContainer(name, imageDir string, entryPoint []string) (*Container, error) {
	cc := &ContainerConfig{
		Name:            name,
		ImageDir:        imageDir,
		ChildEntryPoint: entryPoint,
		Cgroups:         make([]cgroup.ICgroup, 0),
	}
	cs := &ContainerState{
		Name:         name,
		UnionMounted: false,
		LifeCycle:    Created,
		ParentPID:    0,
		ChildPID:     0,
		IPNet:        nil,
	}

	err := common.ErrTag("new created container",
		fs.CreateContainerDir(name),
		fs.UnionMountForContainer(name, imageDir),
	)
	if err != nil {
		RemoveContainerForce(name)
		return nil, err
	}
	cs.UnionMounted = true

	err = common.ErrTag("new created container",
		cc.Save(), cs.Save())
	return &Container{
		Config: cc,
		State:  cs,
	}, err
}

// NewContainerFromDisk 从磁盘上加载容器配置和状态
// 注意：调用前需要确保容器配置和状态文件存在
func NewContainerFromDisk(name string) (*Container, error) {
	cc := &ContainerConfig{Name: name}
	if err := cc.Load(); err != nil {
		return nil, err
	}
	cs := &ContainerState{Name: name}
	if err := cs.Load(); err != nil {
		return nil, err
	}
	return &Container{
		Config: cc,
		State:  cs,
	}, nil
}

// InitHostConfig 初始化宿主机配置：创建配置根目录、创建网桥、创建IP池
func InitHostConfig() error {
	return common.Err(common.ErrGroup(
		os.MkdirAll(config.ContainerMountDir, 0755),
		os.MkdirAll(config.ContainerWorkDir, 0755),
		os.MkdirAll(config.ContainerCOWDir, 0755),
		os.MkdirAll(config.ContainerConfigDir, 0755),

		network.InitBridgeAndIPPool(),
	))
}

// ListContainers 列出所有容器
func ListContainers() ([]*Container, error) {
	containers := make([]*Container, 0)

	dirs, err := os.ReadDir(config.ContainerConfigDir)
	if err != nil {
		return containers, err
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		container, err := NewContainerFromDisk(dir.Name())
		if err != nil {
			return containers, err
		}
		containers = append(containers, container)
	}

	return containers, nil
}

// RemoveContainerForce 强制删除容器
// 注意：该删除方法不涉及删除cgroups
func RemoveContainerForce(name string) error {
	return common.ErrTag("clear container",
		fs.UnionUnmountForContainer(name),
		fs.DeleteContainerDir(name),
	)
}
