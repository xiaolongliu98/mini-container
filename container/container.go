package container

import (
	"fmt"
	"mini-container/common"
	"mini-container/config"
	"mini-container/internal/fs"
	"mini-container/internal/network"
	"net"
	"path/filepath"
)

const (
	StateName  = "state.json"
	ConfigName = "config.json"
)

type LifeCycle string

const (
	Unknown LifeCycle = "unknown"
	Created LifeCycle = "created"
	Running LifeCycle = "running"
	Stopped LifeCycle = "stopped"
)

// ContainerConfig 容器配置
// ~/.mini-container/config/<container name>/config.json
type ContainerConfig struct {
	Name            string   `json:"name"`
	ImageDir        string   `json:"imageDir"`
	ChildEntryPoint []string `json:"childEntryPoint"`
}

func (cc *ContainerConfig) Load() error {
	return common.ReadJSON(filepath.Join(config.ContainerConfigDir, cc.Name, ConfigName), cc)
}

func (cc *ContainerConfig) Save() error {
	return common.WriteJSON(filepath.Join(config.ContainerConfigDir, cc.Name, ConfigName), cc)
}

// ContainerState 容器状态
// ~/.mini-container/config/<container name>/state.json
type ContainerState struct {
	Name         string     `json:"name"`
	UnionMounted bool       `json:"unionMounted"`
	LifeCycle    LifeCycle  `json:"lifeCycle"`
	ParentPID    int        `json:"parentPID"`
	ChildPID     int        `json:"childPID"`
	IPNet        *net.IPNet `json:"ipNet"` // x.x.x.x/x, 如果不为nil，表示已经分配了ip，如果在stopped状态需要释放ip
}

func (cs *ContainerState) Load() error {
	return common.ReadJSON(filepath.Join(config.ContainerConfigDir, cs.Name, StateName), cs)
}

func (cs *ContainerState) Save() error {
	return common.WriteJSON(filepath.Join(config.ContainerConfigDir, cs.Name, StateName), cs)
}

type Container struct {
	Config *ContainerConfig
	State  *ContainerState
}

// SetRunning 设置容器为运行状态
// 调用该方法前你需要保证child进程已经启动
func (c *Container) SetRunning(parentPID, childPID int) error {
	c.State.LifeCycle = Running
	c.State.ParentPID = parentPID
	c.State.ChildPID = childPID
	return c.State.Save()
}

// Kill 强制停止容器
func (c *Container) Kill() error {
	if !c.IsRunning() {
		return nil
	}

	err := common.ErrTag("kill",
		common.ErrTag("child process", common.KillProc(c.State.ChildPID)),
		common.ErrTag("parent process", common.KillProc(c.State.ParentPID)),
	)
	if err != nil {
		return err
	}

	return c.SetStopped()
}

func (c *Container) SetStopped() error {
	c.State.LifeCycle = Stopped
	c.State.ParentPID = 0
	c.State.ChildPID = 0

	// 释放ip，释放失败不影响运行停止
	if c.State.IPNet != nil {
		if !common.ErrLog("release container ip",
			network.ReleaseNetworkForContainer(c.State.IPNet.String())) {
			c.State.IPNet = nil
		}
	}
	return c.State.Save()
}

// ConfigChildNetworkForParent 配置容器的网络
// 调用该方法前你需要保证child进程已经启动，并且已经调用 SetRunning
func (c *Container) ConfigChildNetworkForParent() error {
	ipNet, err := network.ConfigNetworkForContainer(c.State.ChildPID)
	if err != nil {
		return err
	}
	c.State.IPNet = ipNet
	return c.State.Save()
}

func (c *Container) ConfigRootfsForChild() error {
	return fs.ChangeRootForContainer(c.Config.Name)
}

// IsRunning 判断容器是否在运行状态
func (c *Container) IsRunning() bool {
	if c.State.LifeCycle != Running {
		return false
	}
	// check pid
	if common.IsExistProc(c.State.ChildPID) {
		return true
	}
	// update
	if err := c.SetStopped(); err != nil {
		common.ErrLog("update container state", err)
	}
	return false
}

// GetLifeCycle 获取容器的生命周期
func (c *Container) GetLifeCycle() LifeCycle {
	switch c.State.LifeCycle {
	case Created:
		return Created
	case Running:
		if c.IsRunning() {
			return Running
		} else {
			return Stopped
		}
	case Stopped:
		return Stopped
	default:
		return Unknown
	}
}

// Remove 删除容器
// 调用该方法前你需要保证容器已经停止
func (c *Container) Remove() error {
	if c.IsRunning() {
		return fmt.Errorf("container %s is running", c.Config.Name)
	}
	return RemoveContainerForce(c.Config.Name)
}

// FixCreatedToStopped 修复容器状态
func (c *Container) FixCreatedToStopped() error {
	if c.State.LifeCycle != Created {
		return fmt.Errorf("container %s is not created", c.Config.Name)
	}

	// 检查容器初始化配置
	// 暂时没有需要检查的

	c.State.LifeCycle = Stopped
	return c.State.Save()
}
