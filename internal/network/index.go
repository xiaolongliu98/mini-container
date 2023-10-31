package network

import (
	"fmt"
	"math/rand"
	"mini-container/config"
	"mini-container/internal/network/bridge"
	"mini-container/internal/network/ippool"
	"net"
)

var (
	IPPool *ippool.IPPool
)

// InitBridgeAndIPPool 初始化宿主机网桥，读取IP Pool文件
func InitBridgeAndIPPool() error {
	var err error
	IPPool, err = ippool.NewFromDiskIfExists(config.IPPoolPath)
	if err != nil {
		return err
	}
	err = IPPool.SetUsed(config.DefaultBridgeIPNet)
	if err != nil {
		return err
	}

	if !bridge.ExistsBridge(config.DefaultBridgeName) {
		err = bridge.CreateBridgeAndSetSNAT(config.DefaultBridgeName, config.DefaultBridgeIPNet)
		if err != nil {
			return err
		}

	}
	return nil
}

// ConfigNetworkForContainer 配置容器IP
// return: allocateIPNet, error
func ConfigNetworkForContainer(pid int) (*net.IPNet, error) {
	// 分配IP
	allocateIPNet, err := IPPool.AllocateIP(config.DefaultBridgeIPNet)
	if err != nil {
		return nil, fmt.Errorf("alloc allocateIPNet fail %s", err)
	}

	// 主机上创建 veth 设备,并连接到网桥上
	randPart := rand.Intn(900) + 100 // 100~999
	vethName := fmt.Sprintf("%d-%d", pid, randPart)

	peerName, err := bridge.CreateVeth(config.DefaultBridgeName, vethName)

	//fmt.Println("[Debug]peerName:", peerName)

	if err != nil {
		return nil, fmt.Errorf("create veth fail err=%s", err)
	}
	// 主机上设置子进程网络命名空间配置
	bridgeIPNet, _ := bridge.ParseIPNet(config.DefaultBridgeIPNet)
	if err := bridge.SetContainerIP(peerName, pid, allocateIPNet.IP, bridgeIPNet); err != nil {
		return nil, fmt.Errorf("SetContainerIP fail err=%s peer-name=%s pid=%d allocateIPNet=%v", err, peerName, pid, allocateIPNet)
	}
	return allocateIPNet, nil
}

// ReleaseNetworkForContainer 释放容器IP配置
func ReleaseNetworkForContainer(ipNetStr string) error {
	return IPPool.ReleaseIPStr(ipNetStr)
}

func ReleaseBridge() error {
	// TODO

	return nil
}
