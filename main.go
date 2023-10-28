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

// run [container name] [image path] [entry point] [args...]
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

// ~ run [container name] [image path] [entry point] [args...]
func parent() {
	log.Printf("RUNNING parent as PID %d\n", os.Getpid())

	var (
		containerName = os.Args[2]
		imageDir      = os.Args[3]
	)
	// STEP 1 初始化br0网桥配置

	// parent start child process
	os.Args[1] = CMDNameChild
	cmd := exec.Command(ProcSelfExe, os.Args[1:]...)

	// STEP 2 设置 child process 的 Namespace 来隔离各种资源
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

	if _, err := common.ErrGroup(
		fs.CreateInstanceDir(containerName),
		fs.UnionMountForInstance(containerName, imageDir),
	); err != nil {
		log.Fatalln("ERROR 1 parent", err)
	}

	// Start 异步启动， Run 同步启动
	if err := cmd.Start(); err != nil {
		log.Println("ERROR 2 parent", err)
	}

	// TODO child进程初始化完毕后，再执行下方
	// TODO 设置Cgroups
	// TODO 设置network

	if err := cmd.Wait(); err != nil {
		log.Println("ERROR 3 parent", err)
	}

	// 删除操作
	if i, err := common.ErrGroupThrough(
		fs.UnionUnmountForInstance(containerName),
		fs.DeleteInstanceDir(containerName),
		// TODO 清空Cgroups目录
	); err != nil {
		log.Printf("ERROR 4.%v parent %s", i, err)
	}
}

// ~ child [container name] [image path] [entry point] [args...]
func child() {
	var (
		containerName = os.Args[2]
		childCMD      = os.Args[4]
	)

	log.Printf("RUNNING %v as PID %d\n", childCMD, os.Getpid())

	// STEP 3: 挂载文件系统 or 隔离文件系统
	common.Must(fs.ChangeRootForInstance(containerName))

	// 注意：syscall.Exec 是替换当前进程，cmd.Run 是创建一个新的进程
	if err := syscall.Exec(childCMD, os.Args[4:], os.Environ()); err != nil {
		log.Println("ERROR 1 child", err)
	}
}
