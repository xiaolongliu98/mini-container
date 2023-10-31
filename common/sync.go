package common

import (
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
