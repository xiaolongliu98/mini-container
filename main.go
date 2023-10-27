package main

import (
	"log"
	common "mini-container/common"
	"os"
	"os/exec"
	"path/filepath"
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

	// 运行命令并检查错误
	if err := cmd.Run(); err != nil {
		log.Fatalln("ERROR parent", err)
	}
}

func child() {
	var (
		childCMD          = os.Args[2]
		rootfs            = "./rootfs"
		oldRootfs         = filepath.Join(rootfs, ".old")
		defaultMountFlags = syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	)

	log.Printf("RUNNING %v as PID %d\n", childCMD, os.Getpid())

	// STEP 2: 挂载文件系统 or 隔离文件系统
	common.Must(
		// 设置私有的 mount namespace，这样在当前 namespace 中的挂载操作不会影响到parent namespace
		syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""),

		// 重新挂载 rootfs，使得 rootfs 成为当前 namespace 的根目录，以下是固定的4个步骤
		syscall.Mount(rootfs, rootfs, "bind", syscall.MS_BIND|syscall.MS_REC, ""),
		os.MkdirAll(oldRootfs, 0700),
		syscall.PivotRoot(rootfs, oldRootfs),
		syscall.Chdir("/"),

		// 挂载 proc 文件系统的目的: 为了让 ps 等命令能够正常运行
		// 但需要注意的是：如果rootfs中不是一个完整的操作系统，那么挂载 proc 文件系统可能会失败
		os.MkdirAll("/proc", 0755),
		syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""),
	)
	// 注意：syscall.Chroot() 只能改变当前进程的根目录，不能改变当前进程所属的 Namespace 的根目录

	// TODO STEP 3: 设置 Cgroup 来限制资源使用
	// cg()

	// 注意：syscall.Exec 是替换当前进程，cmd.Run 是创建一个新的进程
	if err := syscall.Exec(childCMD, os.Args[2:], os.Environ()); err != nil {
		log.Fatalln("ERROR child", err)
	}
}
