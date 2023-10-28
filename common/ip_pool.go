package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

type IPPool struct {
	//cidr -> bitmap
	m map[string]*Bitmap
}

func NewIPPool() *IPPool {
	return &IPPool{m: make(map[string]*Bitmap)}
}

func SaveIPPool(pool *IPPool, dir, name string) error {
	jsonBytes, _ := json.Marshal(pool.m)
	return os.WriteFile(filepath.Join(dir, name), jsonBytes, 0644)
}

func LoadIPPool(dir, name string) (*IPPool, error) {
	jsonBytes, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, name))
	if err != nil {
		return nil, err
	}

	pool := NewIPPool()
	err = json.Unmarshal(jsonBytes, &pool.m)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// AllocateIP allocate an ip from the pool
// ipNetStr: x.x.x.x/x
// return: IP likes x.x.x.x
func (p *IPPool) AllocateIP(ipNetStr string) (string, error) {
	// ip: 192.168.0.1/24
	_, ipNet, err := net.ParseCIDR(ipNetStr)
	if err != nil {
		return "", err
	}

	ipNetStr = ipNet.String()

	bm, ok := p.m[ipNetStr]
	if !ok {
		ones, _ := ipNet.Mask.Size()
		validIPs := 1 << uint(32-ones)

		bm = NewBitmap(validIPs)
		p.m[ipNetStr] = bm
	}

	if bm.Ones() >= bm.Cap()-2 {
		return "", errors.New("no available ip")
	}

	unsetPos := bm.GetFirstUnset(1)
	_ = bm.Set(unsetPos) // no error

	ip := ipNet.IP.To4()
	ip1Uint32 := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
	ip2Uint32 := uint32(unsetPos)

	ipUint32 := ip1Uint32 | ip2Uint32
	ip[0] = byte(ipUint32 >> 24)
	ip[1] = byte(ipUint32 >> 16)
	ip[2] = byte(ipUint32 >> 8)
	ip[3] = byte(ipUint32)

	return ip.String(), nil
}

// ReleaseIP release an ip to the pool
// ipNetStr: x.x.x.x/x
func (p *IPPool) ReleaseIP(ipNetStr string) error {
	ip, ipNet, err := net.ParseCIDR(ipNetStr)
	if err != nil {
		return err
	}

	ipNetStr = ipNet.String()
	ip = ip.To4()

	bm, ok := p.m[ipNetStr]
	if !ok {
		return nil
	}

	// get ip pos
	ones, _ := ipNet.Mask.Size()
	ipUint32 := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
	pos := ipUint32 & ((1 << uint(32-ones)) - 1)

	bm.Unset(int(pos))
	return nil
}

// IsAvailable check if an ip is available
// ipNetStr: x.x.x.x/x
func (p *IPPool) IsAvailable(ipNetStr string) bool {
	ip, ipNet, err := net.ParseCIDR(ipNetStr)
	if err != nil {
		return false
	}

	ipNetStr = ipNet.String()
	ip = ip.To4()

	// get ip pos
	ones, _ := ipNet.Mask.Size()
	ipUint32 := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
	pos := ipUint32 & ((1 << uint(32-ones)) - 1)
	validIPs := 1 << uint(32-ones)
	if pos == 0 || pos == uint32(validIPs-1) {
		return false
	}

	bm, ok := p.m[ipNetStr]
	if !ok {
		return true
	}
	return !bm.Get(int(pos))
}
