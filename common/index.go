package common

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

func TrimError(err error) string {
	return strings.Trim(strings.TrimLeft(err.Error(), "ERROR"), " ")
}

func Must(err ...error) {

	for i, e := range err {
		if e != nil {
			if len(err) == 1 {
				fmt.Printf("ERROR %s\n", TrimError(e))
			} else {
				fmt.Printf("ERROR [%d]%s\n", i, TrimError(e))
			}
			os.Exit(1)
		}
	}
}

func MustLog(errTag string, err ...error) {
	for i, e := range err {
		if e != nil {
			if len(err) == 1 {
				fmt.Printf("ERROR %s -> %s\n", errTag, TrimError(e))
			} else {
				fmt.Printf("ERROR [%d]%s -> %s\n", i, errTag, TrimError(e))
			}
			os.Exit(1)
		}
	}
}

func ErrLog(errTag string, err ...error) bool {
	isErr := false

	for i, e := range err {
		if e != nil {
			if len(err) == 1 {
				fmt.Printf("ERROR %s -> %s\n", errTag, TrimError(e))
			} else {
				fmt.Printf("ERROR [%d]%s -> %s\n", i, errTag, TrimError(e))
			}
			isErr = true
		}
	}

	return isErr
}

func ErrTag(errTag string, err ...error) error {
	for i, e := range err {
		if e != nil {
			if len(err) == 1 {
				return fmt.Errorf("ERROR %s -> %s", errTag, TrimError(e))
			} else {
				return fmt.Errorf("ERROR [%d]%s -> %s", i, errTag, TrimError(e))
			}
		}
	}
	return nil
}

func Err(rets ...any) error {
	for i := len(rets) - 1; i >= 0; i-- {
		if err, ok := rets[i].(error); ok {
			return err
		}
	}
	return nil
}

func ErrGroup(err ...error) (int, error) {
	for i, e := range err {
		if e != nil {
			return i, e
		}
	}
	return -1, nil
}

func ErrGroupCount(err ...error) (int, error) {
	var targetErr error = nil
	var count int = -1

	for _, e := range err {
		if e != nil {
			count++
		}

		if e != nil && targetErr == nil {
			targetErr = e
		}
	}

	return count, targetErr
}

// IsExistPath 判断所给路径文件/文件夹是否存在
func IsExistPath(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// IsExistProc 判断所给PID进程是否存在
func IsExistProc(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Sending signal 0 to a pid will not kill the process.
	// If the system successfully delivers the signal, then the process is alive.
	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}

	// err is not nil; if the error is "process already finished" then the process doesn't exist
	if err.Error() == "os: process already finished" {
		return false
	}

	// Otherwise, there might be other reasons for the error (like permission denied),
	// which means the process might still exist.
	return true
}

// KillProc 强制杀死进程
func KillProc(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}
