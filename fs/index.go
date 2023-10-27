package fs

import "syscall"

// 		syscall.Mount("rootfs", "rootfs", "", syscall.MS_BIND, ""),
//		os.MkdirAll("rootfs/oldrootfs", 0700),
//		syscall.PivotRoot("rootfs", "rootfs/oldrootfs"),
//		os.Chdir("/"),

func ChRoot(root string) error {
	return syscall.Chroot(root)
}
