package common

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func Signal(pid int) error {
	return syscall.Kill(pid, syscall.SIGUSR2)
}

func WaitSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR2)
	<-sigs
}

// NewWaitSignalChannel 用于创建一个阻塞等待信号的channel
// return: wait function, call it to wait signal
func NewWaitSignalChannel() func() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR2)
	return func() {
		<-sigs
	}
}

func WriteJSON(name string, obj any) error {
	data, _ := json.Marshal(obj)
	return os.WriteFile(name, data, os.ModePerm)
}

func WriteJSONSync(name string, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

func ReadJSON(name string, objPtr any) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, objPtr)
}
