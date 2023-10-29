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
