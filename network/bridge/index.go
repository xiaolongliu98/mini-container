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

func truncate(maxLen int, str string) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen]
}

// createBridge 使用给定的名称和接口IP创建桥接。
func createBridge(bridgeName string, interfaceIPNet *net.IPNet) error {
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	br := &netlink.Bridge{LinkAttrs: la}

	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge creation failed for bridge %s: %s", bridgeName, err)
	}

	addr := &netlink.Addr{IPNet: interfaceIPNet, Peer: interfaceIPNet, Label: "", Flags: 0, Scope: 0}
	if err := netlink.AddrAdd(br, addr); err != nil {
		return fmt.Errorf("bridge add addr fail %s", err)
	}

	if err := netlink.LinkSetUp(br); err != nil {
		return fmt.Errorf("error enabling interface for %s: %v", bridgeName, err)
	}
	return nil
}

// ParseIPNet "192.167.0.100/24" -> *net.IPNet
func ParseIPNet(ipNetStr string) (*net.IPNet, error) {
	ip, ipNet, err := net.ParseCIDR(ipNetStr)
	if err != nil {
		return nil, err
	}
	ipNet.IP = ip
	return ipNet, nil
}

// setSNAT 在 Linux 网络配置中，SNAT（Source Network Address Translation）
// 主要用于改变出站数据包的源 IP 地址。对于 Linux bridge 设备，设置 SNAT 有以下可能的用途：
//
//	1 -> 提供网络隔离：在虚拟化或容器化环境中，通常会创建多个网络命名空间以隔离不同的工作负载。
//	  这些网络命名空间可能会连接到同一个 Linux bridge。通过在 bridge 上设置 SNAT，
//	  可以确保每个网络命名空间的出站流量都具有唯一的源 IP 地址，从而避免 IP 地址冲突。
//	2 -> 实现 IP 伪装：在某些情况下，你可能希望隐藏内部网络的真实 IP 地址。通过在 bridge
//	  上设置 SNAT，可以将所有出站流量的源 IP 地址改为 bridge 的 IP 地址，从而隐藏内部网络的真实 IP 地址。
//	3 -> 提供网络访问：在某些情况下，你可能希望让没有公网 IP 地址的内部网络能够访问公网。
//	  通过在连接到公网的 bridge 上设置 SNAT，可以让内部网络的出站流量看起来像是来自 bridge 的 IP 地址，
//	  从而实现内部网络访问公网。
//
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
func setInterfaceIPNet(name string, ipNetStr string) error {
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
	ipNet, err := ParseIPNet(ipNetStr)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{IPNet: ipNet, Peer: ipNet, Label: "", Flags: 0, Scope: 0}
	return netlink.AddrAdd(iface, addr)
}

// CreateBridgeAndSetSNAT 创建网桥并设置SNAT规则
// bridgeName string：网桥名称，长度不能超过15个字符
// ipNetStr string：网桥的IP地址，格式为：x.x.x.x/x
// 注意：使用前请检查bridgeName是否已经存在，系统重启后需要重新建立网桥配置
func CreateBridgeAndSetSNAT(bridgeName string, ipNetStr string) error {
	bridgeName = truncate(15, bridgeName)

	// 创建网桥
	interfaceIP, err := ParseIPNet(ipNetStr)
	if err != nil {
		return fmt.Errorf("ParseIPNet err=%s", err)
	}

	err = createBridge(bridgeName, interfaceIP)
	if err != nil {
		return fmt.Errorf("createBridge err=%s", err)
	}
	_, subnet, _ := net.ParseCIDR(ipNetStr)
	return setSNAT(bridgeName, subnet)
}

func ExistsBridge(bridgeName string) bool {
	bridgeName = truncate(15, bridgeName)
	_, err := net.InterfaceByName(bridgeName) // 原生实现
	//_, err := netlink.LinkByName(bridgeName) // 参考实现
	return err == nil
}

// CreateVeth 创建veth设备
// bridgeName string：网桥名称，长度不能超过15个字符
// vethName string：veth设备名称，长度不能超过10个字符
// return: veth peer name
// 注意：使用前请检查bridgeName是否已经存在
func CreateVeth(bridgeName, vethName string) (string, error) {
	bridgeName = truncate(15, bridgeName)
	vethName = truncate(10, vethName)

	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return "", fmt.Errorf("link by name fail err=%s", err)
	}

	la := netlink.NewLinkAttrs()
	la.Name = "veth-" + vethName
	la.MasterIndex = br.Attrs().Index

	// 创建veth设备
	vethLink := &netlink.Veth{
		LinkAttrs: la,
		PeerName:  truncate(15, "peer-"+vethName),
	}

	if err := netlink.LinkAdd(vethLink); err != nil {
		return "", fmt.Errorf("veth creation failed for bridge %s: %s", bridgeName, err)
	}

	if err := netlink.LinkSetUp(vethLink); err != nil {
		return "", fmt.Errorf("error enabling interface for %s: %v", vethName, err)
	}

	return vethLink.PeerName, nil
}

func SetContainerIP(peerName string, pid int, containerIP net.IP, gateway *net.IPNet) error {
	peerLink, err := netlink.LinkByName(peerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}
	loLink, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}

	// 进入容器的网络命名空间
	recoverFunc, err := enterNetworkNameSpace(&peerLink, pid)
	defer recoverFunc()
	if err != nil {
		return fmt.Errorf("enterNetworkNameSpace fail err=%s", err)
	}

	peerIPNet := &net.IPNet{
		IP:   containerIP,
		Mask: gateway.Mask,
	}
	if err := setInterfaceIPNet(peerName, peerIPNet.String()); err != nil {
		return fmt.Errorf("%v,%s", containerIP, err)
	}
	if err := netlink.LinkSetUp(peerLink); err != nil {
		return fmt.Errorf("netlink.LinkSetUp fail  name=%s err=%s", peerName, err)
	}
	if err := netlink.LinkSetUp(loLink); err != nil {
		return fmt.Errorf("netlink.LinkSetUp fail  name=%s err=%s", peerName, err)
	}

	// 为容器设置默认路由
	// LinkIndex: 是网卡的索引，这里是Veth peer的索引
	// Gw: 网关地址
	// Dst: 目标地址
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        gateway.IP,
		Dst:       cidr,
	}
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("router add fail %s", err)
	}

	return nil
}
