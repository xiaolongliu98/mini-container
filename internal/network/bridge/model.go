package bridge

import "net"

type BridgeConfig struct {
	Name   string     `json:"name"`
	IPNet  *net.IPNet `json:"ipNet"`
	Subnet *net.IPNet `json:"subnet"`
}
