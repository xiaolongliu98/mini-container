package fs

import (
	"errors"
	"mini-container/common"
	"path/filepath"
)

var (
	StateName = "state.json"
)

type LifeCycle string

const (
	Created LifeCycle = "created"
	Running LifeCycle = "running"
	Stopped LifeCycle = "stopped"
)

type InstanceState struct {
	Name         string    `json:"name"`
	UnionMounted bool      `json:"unionMounted"`
	ImageDir     string    `json:"imageDir"`
	LifeCycle    LifeCycle `json:"lifeCycle"`
}

func NewInstanceStateFromDisk(name string) (*InstanceState, error) {
	state := &InstanceState{}
	if err := state.Load(name); err != nil {
		return nil, err
	}
	return state, nil
}

func NewCreatedInstanceState(name string) *InstanceState {
	return &InstanceState{
		Name:         name,
		UnionMounted: false,
		ImageDir:     "",
		LifeCycle:    Created,
	}
}

func (s *InstanceState) ToRunning() error {
	s.LifeCycle = Running
	return nil
}

func (s *InstanceState) ToStopped() error {
	if s.LifeCycle != Running {
		return errors.New("instance is not running")
	}
	s.LifeCycle = Stopped
	return nil
}

func (s *InstanceState) Load(name string) error {
	statePath := filepath.Join(InstanceStateDir, name, StateName)
	if err := common.ReadJSON(statePath, s); err != nil {
		return err
	}
	return nil
}

func (s *InstanceState) Save() error {
	statePath := filepath.Join(InstanceStateDir, s.Name, StateName)
	return common.WriteJSON(statePath, s)
}

func (s *InstanceState) SetMount(imageDir string) {
	s.ImageDir = imageDir
	s.UnionMounted = true
}
