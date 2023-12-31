package fs

import (
	"fmt"
	"mini-container/common"
	"mini-container/config"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// OverlayFS 是 Linux 内核中的一种联合文件系统（Union filesystem），
// 它可以将多个不同的文件系统（称为层）叠加在一起，形成一个单一、统一的文件系统视图。
// OverlayFS 主要由两层组成：下层（lower layer）和上层（upper layer）。
// 下层是只读的，而上层是可写的。OverlayFS 主要用于为 Docker 和其他容器技术提供存储解决方案。
// 在 Docker 中，每个镜像层和容器层都可以被视为一个 OverlayFS 层。
// 当启动一个容器时，Docker 会将所有相关的镜像层和一个新的可写容器层叠加在一起，
// 形成一个单一的文件系统视图。这使得容器可以看到一个完整的文件系统，但实际上它只修改了最上面的那一层。
// 你可以使用 `mount` 命令来手动挂载一个 OverlayFS。以下是一个例子：
//
// ```bash
// mount -t overlay overlay -o lowerdir=/lower,upperdir=/upper,workdir=/work /merged
// ```
//
// 在这个例子中，`/lower` 是下层目录，`/upper` 是上层目录，`/work` 是
// 工作目录（用于存储一些必要的元数据），`/merged` 是挂载点。挂载后，
// 你可以在 `/merged` 中看到下层和上层的文件，任何对 `/merged` 的修改都会在上层反映出来。
// 请注意，要使用 OverlayFS，你需要一个支持 OverlayFS 的 Linux 内核（3.18 或更高版本）。

func CreateContainerDir(name string) error {
	if ExistsContainerDir(name) {
		return nil
	}
	return common.Err(common.ErrGroup(
		os.MkdirAll(filepath.Join(config.ContainerMountDir, name), 0755),
		os.MkdirAll(filepath.Join(config.ContainerWorkDir, name), 0755),
		os.MkdirAll(filepath.Join(config.ContainerCOWDir, name), 0755),
		os.MkdirAll(filepath.Join(config.ContainerConfigDir, name), 0755),
	))
}

func DeleteContainerDir(name string) error {
	return common.Err(common.ErrGroup(
		os.RemoveAll(filepath.Join(config.ContainerMountDir, name)),
		os.RemoveAll(filepath.Join(config.ContainerWorkDir, name)),
		os.RemoveAll(filepath.Join(config.ContainerCOWDir, name)),
		os.RemoveAll(filepath.Join(config.ContainerConfigDir, name)),
	))
}

// UnionMountForContainer 联合挂载镜像层和容器层
// 注意：挂载前，你需要先调用CreateContainerDir保证实例目录存在
func UnionMountForContainer(name, imageDir string) error {
	mntDir := filepath.Join(config.ContainerMountDir, name)
	wordDir := filepath.Join(config.ContainerWorkDir, name)
	cowDir := filepath.Join(config.ContainerCOWDir, name)

	imageDir, err := filepath.Abs(imageDir)
	if err != nil {
		return err
	}

	if err := syscall.Mount("overlay", mntDir, "overlay", 0,
		fmt.Sprintf("upperdir=%s,lowerdir=%s,workdir=%s",
			cowDir, imageDir, wordDir)); err != nil {
		return err
	}
	return nil
}

func ExistsContainerDir(name string) bool {
	mntDir := filepath.Join(config.ContainerMountDir, name)
	return common.IsExistPath(mntDir)
}

// UnionUnmountForContainer 取消联合挂载
// 注意：取消挂载后，你需要调用DeleteContainerDir删除实例目录
func UnionUnmountForContainer(name string) error {
	// 和syscall.Unmount差距在于exec.Command是启动新的进程，而syscall.Unmount是在当前进程中执行
	return common.Err(common.ErrGroup(
		common.Err(exec.Command("umount", filepath.Join(config.ContainerMountDir, name)).CombinedOutput()),
		common.Err(exec.Command("umount", filepath.Join(config.ContainerWorkDir, name)).CombinedOutput()),
		common.Err(exec.Command("umount", filepath.Join(config.ContainerCOWDir, name)).CombinedOutput()),
	))
}

// ChangeRootForContainer ChangeRoot的封装，用于将 rootfs 切换到指定的目录，同时将 oldRootfs 作为挂载点
// 注意：需要在child中执行
func ChangeRootForContainer(name string) error {
	rootfs := filepath.Join(config.ContainerMountDir, name)
	return ChangeRoot(rootfs)
}

// ChangeRoot 用于将 rootfs 切换到指定的目录，同时将 oldRootfs 作为挂载点
// 注意：需要在child中执行
func ChangeRoot(rootfs string) error {
	oldRootfs := filepath.Join(rootfs, config.OldRootfsName)
	return common.Err(common.ErrGroup(
		// 设置私有的 mount namespace，这样在当前 namespace 中的挂载操作不会影响到parent namespace
		syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""),

		// 重新挂载 rootfs，使得 rootfs 成为当前 namespace 的根目录，以下是固定的4个步骤
		syscall.Mount(rootfs, rootfs, "bind", syscall.MS_BIND|syscall.MS_REC, ""),
		os.MkdirAll(oldRootfs, 0700),
		syscall.PivotRoot(rootfs, oldRootfs),
		syscall.Chdir("/"),

		os.MkdirAll("/proc", 0700),
		// 重新挂载 proc 文件系统，使得当前 namespace 中可以访问 proc 文件系统
		MountProc(),
	))
	// 注意：syscall.Chroot() 只能改变当前进程的根目录，不能改变当前进程所属的 Namespace 的根目录
}

// MountProc 用于挂载 proc 文件系统到指定的目录
// 注意：需要在child namespace中执行，且需要先执行 ChangeRoot
// 注意：如果rootfs中不是一个完整的操作系统，那么挂载 proc 文件系统
// 可能会失败，提前创建好 /proc 目录可以避免这个错误，但并不解决问题
func MountProc() error {
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	return syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
}
