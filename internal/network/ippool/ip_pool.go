package ippool

import (
	"errors"
	"fmt"
	"mini-container/common"
	"net"
)

// TODO 通过文件方式进行加锁

type IPPool struct {
	//subnetStr -> bitmap
	m    map[string]*common.Bitmap
	path string
}

func New(path string) (*IPPool, error) {
	pool := &IPPool{
		m:    make(map[string]*common.Bitmap),
		path: path,
	}

	if !common.IsExistPath(path) {
		fmt.Println("[DEBUG New]", pool.m)
		return pool, pool.save()
	}

	return pool, nil
}

// load
func (p *IPPool) load() error {
	if !common.IsExistPath(p.path) {
		return nil
	}

	err := common.ReadJSON(p.path, &(p.m))
	fmt.Println("[DEBUG load]", p.m)
	return err
}

func (p *IPPool) save() error {
	fmt.Println("[DEBUG save]", p.m)
	return common.WriteJSONSync(p.path, p.m)
}

// AllocateIP allocate an ip from the pool
// ipNetStr: x.x.x.x/x
func (p *IPPool) AllocateIP(subnetStr string) (*net.IPNet, error) {
	if err := p.load(); err != nil {
		return nil, err
	}
	fmt.Println("[DEBUG AllocateIP]", p.m)
	// ip: 192.168.0.1/24
	_, ipNet, err := net.ParseCIDR(subnetStr) // ipNet: 192.168.0.0/24
	if err != nil {
		return nil, err
	}

	subnetStr = ipNet.String()

	bm, ok := p.m[subnetStr]
	if !ok {
		ones, _ := ipNet.Mask.Size()
		validIPs := 1 << uint(32-ones)

		bm = common.NewBitmap(validIPs)
		p.m[subnetStr] = bm
	}

	if bm.Ones() >= bm.Cap()-2 {
		return nil, errors.New("no available IP")
	}

	unsetPos := bm.GetFirstUnset(1)
	_ = bm.Set(unsetPos) // no error

	ip := ipNet.IP.To4() // 192.168.0.0
	ip1Uint32 := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
	ip2Uint32 := uint32(unsetPos) // 0.0.0.2

	ipUint32 := ip1Uint32 | ip2Uint32
	ip[0] = byte(ipUint32 >> 24)
	ip[1] = byte(ipUint32 >> 16)
	ip[2] = byte(ipUint32 >> 8)
	ip[3] = byte(ipUint32)

	ipNet.IP = ip
	return ipNet, p.save()
}

// ReleaseIPStr release an ip to the pool
// ipNetStr: x.x.x.x/x
func (p *IPPool) ReleaseIPStr(ipNetStr string) error {
	if err := p.load(); err != nil {
		return err
	}
	fmt.Println("[DEBUG ReleaseIPStr]", p.m)
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
	return p.save()
}

// ReleaseIP release an ip to the pool
func (p *IPPool) ReleaseIP(ipNet *net.IPNet) error {
	return p.ReleaseIPStr(ipNet.String())
}

// IsAvailable check if an ip is available
// ipNetStr: x.x.x.x/x
func (p *IPPool) IsAvailable(ipNetStr string) (bool, error) {
	if err := p.load(); err != nil {
		return false, err
	}
	fmt.Println("[DEBUG IsAvailable]", p.m)
	ip, ipNet, err := net.ParseCIDR(ipNetStr)
	if err != nil {
		return false, err
	}

	ipNetStr = ipNet.String()
	ip = ip.To4()

	// get ip pos
	ones, _ := ipNet.Mask.Size()
	ipUint32 := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
	pos := ipUint32 & ((1 << uint(32-ones)) - 1)
	validIPs := 1 << uint(32-ones)
	if pos == 0 || pos == uint32(validIPs-1) {
		return false, nil
	}

	bm, ok := p.m[ipNetStr]
	if !ok {
		return true, nil
	}
	return !bm.Get(int(pos)), nil
}

// SetUsed set an ip to used
// ipNetStr: x.x.x.x/x
func (p *IPPool) SetUsed(ipNetStr string) error {
	if err := p.load(); err != nil {
		return err
	}
	fmt.Println("[DEBUG SetUsed]", p.m)
	ip, ipNet, err := net.ParseCIDR(ipNetStr)
	if err != nil {
		return err
	}

	subnetStr := ipNet.String()
	ip = ip.To4()
	ones, _ := ipNet.Mask.Size()

	bm, ok := p.m[ipNetStr]
	if !ok {

		validIPs := 1 << uint(32-ones)
		bm = common.NewBitmap(validIPs)
		p.m[subnetStr] = bm
	}

	// get ip pos
	ipUint32 := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
	pos := ipUint32 & ((1 << uint(32-ones)) - 1)

	if err = bm.Set(int(pos)); err != nil {
		return err
	}
	fmt.Println("[DEBUG SetUsed2]", p.m)
	return p.save()
}
