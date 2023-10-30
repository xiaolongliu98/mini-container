package main

import (
	"fmt"
	"log"
	common "mini-container/common"
	"mini-container/fs"
	"mini-container/network"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	ProcSelfExe = "/proc/self/exe"

	CMDNameParent = "run"
	CMDNameChild  = "child"
	CMDNameRemove = "rm"
	CMDNameList   = "ls"
)

// run [container name] [image path] [entry point] [args...]
func main() {
	switch os.Args[1] {
	case CMDNameParent:
		parent()
	case CMDNameChild:
		child()
	case CMDNameRemove:
		remove()
	case CMDNameList:
		list()
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

	if fs.ExistsInstance(containerName) {
		common.MustLog("parent 1", fs.SetInstanceRunning(containerName))

	} else {
		common.MustLog("parent 2",
			fs.CreateInstanceDirAndState(containerName),
			fs.SetInstanceRunning(containerName),
			fs.UnionMountForInstance(containerName, imageDir),
		)
	}
	// STEP 1 初始化mini-ctr0网桥配置
	common.MustLog("parent 0", network.Init())

	// parent start child process
	os.Args[1] = CMDNameChild
	cmd := exec.Command(ProcSelfExe, os.Args[1:]...) // equivalent: ~ child [container name] [image path] [entry point] [args...]

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

	// 提前创建信号channel，防止子进程启动完毕后，父进程还没准备好channel阻塞接收
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGUSR2)

	// Start 异步启动， Run 同步启动
	common.ErrLog("parent 3", cmd.Start())

	// TODO child进程初始化完毕后，再执行下方
	// TODO 设置Cgroups
	// 设置network
	// 等待子进程启动完毕
	<-signalCh
	common.MustLog("parent 4",
		network.ConfigNetworkForInstance(cmd.Process.Pid),
		common.Signal(cmd.Process.Pid),
	)

	log.Printf("RUNNING child as PID %d\n", cmd.Process.Pid)

	common.ErrLog("parent 5", cmd.Wait())
	// 清理工作在rm中
	common.ErrLog("parent 6", fs.SetInstanceStopped(containerName))
}

// ~ child [container name] [image path] [entry point] [args...]
func child() {
	// 通知父进程，子进程初始化完毕，可以进行网络配置
	time.Sleep(1 * time.Millisecond) // 防止父进程还没准备好channel阻塞接收
	common.MustLog("child 0", common.Signal(syscall.Getppid()))
	// 等待父进程通知网络配置完毕
	common.WaitSignal()

	var (
		containerName = os.Args[2]
		childCMD      = os.Args[4]
	)

	// STEP 3: 挂载文件系统 or 隔离文件系统
	common.MustLog("child 1", fs.ChangeRootForInstance(containerName))

	// 注意：syscall.Exec 是替换当前进程，cmd.Run 是创建一个新的进程
	common.MustLog("child 2", syscall.Exec(childCMD, os.Args[4:], os.Environ()))
}

// ~ ls
func list() {
	instances, err := fs.ListInstance()
	common.MustLog("list 0", err)

	// Name	Image Status
	fmt.Printf("%v\t%v\t\t\t%v\n", "Name", "Image", "Status")
	for _, instance := range instances {
		fmt.Printf("%v\t%v\t\t\t%v\n", instance.Name, instance.ImageDir, instance.LifeCycle)
	}
}

// ~ rm [container name]
func remove() {
	var (
		containerName = os.Args[2]
	)

	if !fs.ExistsInstance(containerName) {
		log.Printf("container %v not found\n", containerName)
		return
	}

	// 删除操作
	if i, err := common.ErrGroup(
		fs.UnionUnmountForInstance(containerName),
		fs.DeleteInstanceDir(containerName),
		// TODO 清空Cgroups目录
	); err != nil && !strings.Contains(err.Error(), "exit status 32") {
		log.Printf("ERROR %v remove %s", i, err)
	}
}
