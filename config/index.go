package config

// Path
const (
	ConfigDir     = "/root/.mini-container"
	OldRootfsName = ".old"

	// ImageDir(lower) + ContainerCOWDir(upper) = ContainerMountDir(merged)
	// ContainerWorkDir(workdir)用来辅助ContainerMountDir

	ContainerMountDir  = ConfigDir + "/mnt"
	ContainerWorkDir   = ConfigDir + "/work"
	ContainerCOWDir    = ConfigDir + "/cow"
	ContainerConfigDir = ConfigDir + "/config"

	IPPoolPath = ConfigDir + "/ip-pool.json"
)

// Network
const (
	DefaultBridgeName  = "mini-ctr0"
	DefaultBridgeIPNet = "192.172.0.1/24" // "192.172.0.2/24"
)
