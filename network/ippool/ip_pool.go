package ippool

import (
	"errors"
	"mini-container/common"
	"net"
)

type IPPool struct {
	//subnetStr -> bitmap
	m map[string]*common.Bitmap
}

func New() *IPPool {
	return &IPPool{m: make(map[string]*common.Bitmap)}
}

func NewFromDisk(path string) (*IPPool, error) {
	p := &IPPool{m: make(map[string]*common.Bitmap)}
	if common.IsExist(path) {
		err := common.ReadJSON(path, &p.m)
		return p, err
	}
	return p, nil
}

func (p *IPPool) Save(path string) error {
	return common.WriteJSON(path, p.m)
}

// AllocateIP allocate an ip from the pool
// ipNetStr: x.x.x.x/x
// return: IP likes x.x.x.x
func (p *IPPool) AllocateIP(subnetStr string) (string, error) {
	// ip: 192.168.0.1/24
	_, ipNet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return "", err
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

// SetUsed set an ip to used
// ipNetStr: x.x.x.x/x
func (p *IPPool) SetUsed(ipNetStr string) error {
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
	return bm.Set(int(pos))
}
