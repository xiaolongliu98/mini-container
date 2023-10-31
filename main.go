package main

import (
	"fmt"
	"log"
	common "mini-container/common"
	"mini-container/config"
	"mini-container/container"
	"os"
	"os/exec"
	"syscall"
)

const (
	ProcSelfExe = "/proc/self/exe"

	CMDNameParent = "run"
	CMDNameChild  = "child"
	CMDNameRemove = "rm"
	CMDNameList   = "ls"
	CMDNameStart  = "start"
	CMDNameStop   = "stop"
	CMDNameClear  = "clear"
	CMDNameHelp1  = "--help"
	CMDNameHelp2  = "-h"

	HelpText = `
# mini-container --help

Commands:
~ run [container name] [image path] [entry point] [args...]		create and start a container
~ start [container name]										start a stopped or created container
~ stop [container name]											stop a running container
~ ls															list containers and their information 
~ rm [container name] 											remove a container
~ clear															remove all containers		
`
)

// run [container name] [image path] [entry point] [args...]
func main() {
	common.MustLog("init host config", container.InitHostConfig())

	switch os.Args[1] {
	case CMDNameParent:
		var (
			containerName = os.Args[2]
			imageDir      = os.Args[3]
			entryPoint    = os.Args[4:]
		)
		run(containerName, imageDir, entryPoint)

	case CMDNameChild:
		var (
			containerName = os.Args[2]
		)
		child(containerName)

	case CMDNameRemove:
		var (
			containerName = os.Args[2]
		)
		remove(containerName)
	case CMDNameClear:
		clearAll()

	case CMDNameList:
		list()
	case CMDNameStart:
		var (
			containerName = os.Args[2]
		)
		start(containerName)

	case CMDNameStop:
		var (
			containerName = os.Args[2]
		)
		stop(containerName)
	case CMDNameHelp1, CMDNameHelp2:
		fmt.Println(HelpText)

	default:
		fmt.Println("unknown command, what do you want?")
		fmt.Println(HelpText)
	}
}

// ~ run [container name] [image path] [entry point] [args...]
func run(containerName, imageDir string, entryPoint []string) {
	log.Printf("RUNNING parent as PID %d\n", os.Getpid())

	if container.ExistsContainer(containerName) {
		fmt.Printf("container %s already exists, you can use `~ rm %s` to remove it\n", containerName, containerName)
		return
	}

	ctr, err := container.NewCreatedContainer(containerName, imageDir, entryPoint)
	common.MustLog("parent new container", err)

	parent(ctr)
}

func parent(ctr *container.Container) {
	// parent start child process
	// equivalent: ~ child [container name]
	cmd := exec.Command(ProcSelfExe, CMDNameChild, ctr.Config.Name)

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
	waitFunc := common.NewWaitSignalChannel()
	// Start 异步启动， Run 同步启动
	common.MustLog("parent start child", cmd.Start())
	// TODO child进程初始化完毕后，再执行下方
	// TODO 设置Cgroups
	// 设置network
	// 等待子进程启动完毕
	waitFunc()

	common.MustLog("parent config child",
		ctr.SetRunning(os.Getpid(), cmd.Process.Pid),
		ctr.ConfigChildNetworkForParent(),
		common.Signal(cmd.Process.Pid),
	)

	log.Printf("RUNNING child as PID %d\n", cmd.Process.Pid)

	common.ErrLog("parent wait", cmd.Wait())
	// 清理工作在rm中
	common.ErrLog("parent stop", ctr.SetStopped())
}

// ~ child [container name]
func child(containerName string) {
	// 通知父进程，子进程初始化完毕，可以进行网络配置
	waitFunc := common.NewWaitSignalChannel()
	common.MustLog("child signal parent", common.Signal(syscall.Getppid()))
	// 等待父进程通知网络配置完毕
	waitFunc()

	ctr, err := container.NewContainerFromDisk(containerName)
	common.MustLog("child load container", err)

	// STEP 3: 挂载文件系统 or 隔离文件系统
	common.MustLog("child config rootfs", ctr.ConfigRootfsForChild())

	// 注意：syscall.Exec 是替换当前进程，cmd.Run 是创建一个新的进程
	args := ctr.Config.ChildEntryPoint
	common.MustLog("child exec entry point", syscall.Exec(args[0], args[0:], os.Environ()))
}

// ~ ls
func list() {
	containers, err := container.ListContainers()
	common.MustLog("list", err)

	// Name	Image Status
	fmt.Printf("%v\t%v\t\t\t%v\t\t%v\t%v\n", "Name", "Image", "Status", "IP", "CPID")
	for _, e := range containers {
		fmt.Printf("%v\t%v\t\t\t%v\t\t%v\t%v\n",
			e.Config.Name, e.Config.ImageDir, e.GetLifeCycle(), e.State.IPNet.String(), e.State.ChildPID)
	}

}

// ~ rm [container name]
func remove(containerName string) {
	if container.ExistsContainer(containerName) {
		ctr, err := container.NewContainerFromDisk(containerName)
		common.MustLog("remove load container", err)
		if ctr.IsRunning() {
			fmt.Printf("container %s is running, you can use `~ stop %s` to stop it\n", containerName, containerName)
			return
		}
	}

	common.ErrLog("remove", container.RemoveContainerForce(containerName))
}

// ~ start [container name]
func start(containerName string) {
	if !container.ExistsContainer(containerName) {
		fmt.Printf("container %s not found, you can use `~ run %s [image path] [entry point] [args...]` to create it\n", containerName, containerName)
		return
	}

	ctr, err := container.NewContainerFromDisk(containerName)
	common.MustLog("start load container", err)

	if ctr.IsRunning() {
		fmt.Printf("container %s is already running\n", containerName)
		return
	}

	if ctr.GetLifeCycle() == container.Created {
		// created -> stopped
		common.MustLog("start fix", ctr.FixCreatedToStopped())
	}

	// stopped -> running
	parent(ctr)
}

// ~ stop [container name]
func stop(containerName string) {
	if !container.ExistsContainer(containerName) {
		fmt.Printf("container %s not found, you can use `~ run %s [image path] [entry point] [args...]` to create it\n", containerName, containerName)
		return
	}

	ctr, err := container.NewContainerFromDisk(containerName)
	common.MustLog("stop load container", err)

	if !ctr.IsRunning() {
		fmt.Printf("container %s is not running\n", containerName)
		return
	}

	common.ErrLog("stop", ctr.Kill())
}

func clearAll() {
	containers, err := container.ListContainers()
	common.MustLog("list", err)
	for _, e := range containers {
		common.ErrLog("kill and remove", e.Kill(), e.Remove())
	}
	common.ErrLog("clear config root", os.RemoveAll(config.ConfigDir))
}
