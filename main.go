package main

import (
	"log"
	common "mini-container/common"
	"mini-container/fs"
	"os"
	"os/exec"
	"syscall"
)

var (
	ProcSelfExe = "/proc/self/exe"

	CMDNameParent = "run"
	CMDNameChild  = "child"
)

func main() {
	switch os.Args[1] {
	case CMDNameParent:
		parent()
	case CMDNameChild:
		child()
	default:
		log.Fatalln("unknown command, what do you want?")
	}
}

func parent() {
	// parent start child process
	os.Args[1] = CMDNameChild
	cmd := exec.Command(ProcSelfExe, os.Args[1:]...)

	// STEP 1: 设置 child process 的 Namespace 来隔离各种资源
	// CLONE_NEWUTS: hostname
	// CLONE_NEWPID: process id
	// CLONE_NEWNS: mount
	// CLONE_NEWUSER: user
	// CLONE_NEWIPC: ipc 主要是消息队列的隔离
	// CLONE_NEWNET: network
	// CLONE_NEWCGROUP: cgroup
	// CLONE_NEWTIME: time
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET,
	}

	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("RUNNING parent as PID %d\n", os.Getpid())
	// 运行命令并检查错误
	if err := cmd.Run(); err != nil {
		log.Fatalln("ERROR parent", err)
	}
}

func child() {
	var (
		childCMD = os.Args[2]
		rootfs   = "./rootfs"
	)

	log.Printf("RUNNING %v as PID %d\n", childCMD, os.Getpid())

	// STEP 2: 挂载文件系统 or 隔离文件系统
	common.Must(fs.ChangeRoot(rootfs))

	// TODO STEP 3: 设置 Cgroup 来限制资源使用
	// cg()

	// 注意：syscall.Exec 是替换当前进程，cmd.Run 是创建一个新的进程
	if err := syscall.Exec(childCMD, os.Args[2:], os.Environ()); err != nil {
		log.Fatalln("ERROR child", err)
	}
}
