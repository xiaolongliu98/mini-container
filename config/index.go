package config

const (
	// ProjName 项目名
	ProjName = "mini-container"
)

// Path
const (
	ConfigDir     = "/root/.mini-container"
	OldRootfsName = ".old"

	// ImageDir(lower) + ContainerCOWDir(upper) = ContainerMountDir(merged)
	// ContainerWorkDir(workdir)用来辅助ContainerMountDir
	// COW: copy-on-write 写时复制

	ContainerMountDir  = ConfigDir + "/mnt"
	ContainerWorkDir   = ConfigDir + "/work"
	ContainerCOWDir    = ConfigDir + "/cow"
	ContainerConfigDir = ConfigDir + "/config"

	IPPoolPath = ConfigDir + "/ip-pool.json"

	CgroupsDir = "/sys/fs/cgroup/"
)

// Network
const (
	DefaultBridgeName  = "mini-ctr0"
	DefaultBridgeIPNet = "192.172.0.1/24"
)
