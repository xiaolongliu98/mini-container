package test

import (
	"fmt"
	"net"
	"testing"
)

func TestAny1(t *testing.T) {
	deferFunc := func() {
		fmt.Println("deferFunc")
	}

	fmt.Println("TestAny A")
	defer deferFunc()
	fmt.Println("TestAny B")
}

// 坑：指针拷贝
func TestAny2(t *testing.T) {

	_, subnet, _ := net.ParseCIDR("192.168.0.1/24")

	fmt.Println(subnet)

	subnetCopy := *subnet
	subnetCopy.IP = net.ParseIP("192.168.0.2")

	fmt.Println(&subnetCopy)
	fmt.Println(subnet)
}
