package bridge

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	BridgeNetworkType = "bridge"
)

func truncate(maxLen int, str string) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen]
}

// createBridge 使用给定的名称和接口IP创建桥接。
func createBridge(networkName string, interfaceIPNet *net.IPNet) (string, error) {
	bridgeName := truncate(15, fmt.Sprintf("br-%s", networkName))
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	br := &netlink.Bridge{LinkAttrs: la}

	if err := netlink.LinkAdd(br); err != nil {
		return "", fmt.Errorf("bridge creation failed for bridge %s: %s", bridgeName, err)
	}

	addr := &netlink.Addr{IPNet: interfaceIPNet, Peer: interfaceIPNet, Label: "", Flags: 0, Scope: 0}
	if err := netlink.AddrAdd(br, addr); err != nil {
		return "", fmt.Errorf("bridge add addr fail %s", err)
	}

	if err := netlink.LinkSetUp(br); err != nil {
		return "", fmt.Errorf("error enabling interface for %s: %v", bridgeName, err)
	}
	return bridgeName, nil
}

// parseIPNet "192.167.0.100/24" -> *net.IPNet
func parseIPNet(ipNetStr string) (*net.IPNet, error) {
	ipNet, err := netlink.ParseIPNet(ipNetStr)
	if err != nil {
		return nil, fmt.Errorf("parse ip fail ip=%s err=%s", ipNetStr, err)
	}
	return ipNet, nil
}

// setSNAT 主要用于设置源网络地址转换（Source Network Address Translation，SNAT）规则，
// 以便正确地路由从特定网桥发出的流量。
// bridgeName string：需要设置SNAT规则的网桥名称。
// subnet *net.IPNet：源IP地址的子网，这通常是与网桥关联的网络的子网。
func setSNAT(bridgeName string, subnet *net.IPNet) error {
	iptablesCMD := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCMD, " ")...)
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("set snat fail %s", err)
	}
	return nil
}

// enterNetworkNameSpace 主要用于将一个网络链接（veth pair的一端）移动到
// 特定的网络命名空间（通常是容器的网络命名空间），并且将当前的执行线程也切换到
// 这个网络命名空间。当函数执行完成后，会恢复到原来的网络命名空间。
// return: recover function 执行它会将当前线程恢复到原来的网络命名空间。
// 注意：无论函数是否执行成功，都需要执行返回的recover function。
func enterNetworkNameSpace(vethLink *netlink.Link, pid int) (func(), error) {
	file, err := os.OpenFile(fmt.Sprintf("/proc/%d/ns/net", pid), os.O_RDONLY, 0)
	if err != nil {
		return func() {}, fmt.Errorf("error get container net namespace, %v", err)
	}
	fd := file.Fd()
	runtime.LockOSThread() // 锁定当前的操作系统线程，以防止被其它的Go运行时调度器调度到其它的操作系统线程。

	recoverFunc := func() {
		// 执行这个匿名函数会解锁操作系统线程，允许Go运行时调度器再次调度它。
		runtime.UnlockOSThread()
		file.Close()
	}

	// 修改 veth peer 另外一端移到容器的namespace中
	if err = netlink.LinkSetNsFd(*vethLink, int(fd)); err != nil {
		return recoverFunc, fmt.Errorf("error set link netns , %v", err)
	}

	// 获取当前的网络namespace
	origns, err := netns.Get()
	if err != nil {
		return recoverFunc, fmt.Errorf("error get current netns, %v", err)
	}

	recoverFunc = func() {
		// 执行这个匿名函数会将当前线程切换回原来的网络命名空间，并解锁操作系统线程，允许Go运行时调度器再次调度它。
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		file.Close()
	}

	// 设置当前线程到新的网络namespace，并在函数执行完成之后再恢复到之前的namespace
	if err = netns.Set(netns.NsHandle(fd)); err != nil {
		return recoverFunc, fmt.Errorf("error set netns, %v", err)
	}
	return recoverFunc, nil
}

// setInterfaceIPNet 用于设置网络接口的IP地址
func setInterfaceIPNet(name string, rawIPNetStr string) error {
	retries := 2
	var iface netlink.Link
	var err error

	for i := 0; i < retries; i++ {
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		fmt.Println(fmt.Errorf("error retrieving new bridge netlink link [ %s ]... retrying", name))
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("abandoning retrieving the new bridge link from netlink, Run [ ip link ] to troubleshoot the error: %v\n", err)
	}
	ipNet, err := netlink.ParseIPNet(rawIPNetStr)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{IPNet: ipNet, Peer: ipNet, Label: "", Flags: 0, Scope: 0}
	return netlink.AddrAdd(iface, addr)
}
