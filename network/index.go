package network

import (
	"fmt"
	"math/rand"
	"mini-container/network/bridge"
	"mini-container/network/ippool"
	"net"
)

const (
	DefaultBridgeName  = "mini-ctr0"
	DefaultBridgeIPNet = "192.172.0.1/24"

	ConfigDir  = "/root/.mini-container"
	IPPoolFile = ConfigDir + "/ippool.json"
)

var (
	IPPool *ippool.IPPool
)

func Init() error {
	var err error
	IPPool, err = ippool.NewFromDisk(IPPoolFile)
	if err != nil {
		return err
	}

	if !bridge.ExistsBridge(DefaultBridgeName) {
		err = bridge.CreateBridgeAndSetSNAT(DefaultBridgeName, DefaultBridgeIPNet)
		if err != nil {
			return err
		}
		err = IPPool.SetUsed(DefaultBridgeIPNet)
		if err != nil {
			return err
		}
		err = IPPool.Save(IPPoolFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func ConfigNetworkForInstance(pid int) error {
	// 分配IP
	ip, err := IPPool.AllocateIP(DefaultBridgeIPNet)
	if err != nil {
		return fmt.Errorf("alloc ip fail %s", err)
	}
	if err := IPPool.Save(IPPoolFile); err != nil {
		return fmt.Errorf("save ip fail %s", err)
	}

	// 主机上创建 veth 设备,并连接到网桥上
	randPart := rand.Intn(9000) + 1000 // 1000~9999
	vethName := fmt.Sprintf("%d-%d", pid, randPart)

	peerName, err := bridge.CreateVeth(DefaultBridgeName, vethName)

	fmt.Println("[Debug]peerName:", peerName)

	if err != nil {
		return fmt.Errorf("create veth fail err=%s", err)
	}
	// 主机上设置子进程网络命名空间配置
	bridgeIPNet, _ := bridge.ParseIPNet(DefaultBridgeIPNet)
	if err := bridge.SetContainerIP(peerName, pid, net.ParseIP(ip), bridgeIPNet); err != nil {
		return fmt.Errorf("SetContainerIP fail err=%s peer-name=%s pid=%d ip=%v", err, peerName, pid, ip)
	}

	return nil
}
