package common

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// TestIPPoolWrite
func TestIPPoolWrite(t *testing.T) {
	var err error
	var ip string
	pool := NewIPPool()

	ip, err = pool.AllocateIP("192.168.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
	ip, err = pool.AllocateIP("192.168.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
	ip, err = pool.AllocateIP("192.168.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)

	err = pool.ReleaseIP("192.168.0.2/24")
	if err != nil {
		t.Fatal(err)
	}

	ip, err = pool.AllocateIP("192.168.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
	ip, err = pool.AllocateIP("192.168.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
	t.Log("----------------------------")

	ip, err = pool.AllocateIP("192.168.3.12/24")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
	ip, err = pool.AllocateIP("192.168.3.12/24")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)

	t.Log("----------------------------")

	ip, err = pool.AllocateIP("192.168.128.255/24")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
	ip, err = pool.AllocateIP("192.168.128.255/24")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)

	err = SaveIPPool(pool, "./", "ip_pool.json")
	if err != nil {
		t.Fatal(err)
	}
}

// TestIPPoolRead
func TestIPPoolRead(t *testing.T) {
	var err error
	pool, err := LoadIPPool("./", "ip_pool.json")
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, pool.IsAvailable("192.168.0.1/24"))
	assert.False(t, pool.IsAvailable("192.168.0.2/24"))
	assert.False(t, pool.IsAvailable("192.168.0.3/24"))
	assert.False(t, pool.IsAvailable("192.168.0.4/24"))
	assert.True(t, pool.IsAvailable("192.168.0.5/24"))

	assert.False(t, pool.IsAvailable("192.168.3.1/24"))
	assert.False(t, pool.IsAvailable("192.168.3.2/24"))
	assert.True(t, pool.IsAvailable("192.168.3.3/24"))

	assert.False(t, pool.IsAvailable("192.168.128.1/24"))
	assert.False(t, pool.IsAvailable("192.168.128.2/24"))
	assert.True(t, pool.IsAvailable("192.168.128.3/24"))

	assert.True(t, pool.IsAvailable("192.168.121.3/24"))
	assert.False(t, pool.IsAvailable("192.168.111.0/24"))
	assert.False(t, pool.IsAvailable("192.168.111.255/24"))
	assert.True(t, pool.IsAvailable("192.168.111.254/24"))
	assert.True(t, pool.IsAvailable("192.168.111.1/24"))

	os.RemoveAll("./ip_pool.json")
}
