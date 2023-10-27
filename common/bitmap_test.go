package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// test bitmap
func TestBitmap(t *testing.T) {
	bm := NewBitmap()

	bm.Set(0)
	bm.Set(1)
	bm.Set(12)
	bm.Set(74)
	bm.Set(456)

	t.Log(bm.String())
	assert.True(t, bm.Get(0) == true)
	assert.True(t, bm.Get(1) == true)
	assert.True(t, bm.Get(12) == true)
	assert.True(t, bm.Get(74) == true)
	assert.True(t, bm.Get(2) == false)
	assert.True(t, bm.Get(13) == false)
	assert.True(t, bm.Get(75) == false)

	bm.Unset(0)
	bm.Unset(1)
	bm.Unset(12)
	bm.Unset(74)
	t.Log(bm.String())
	assert.True(t, bm.Get(0) == false)
	assert.True(t, bm.Get(1) == false)
	assert.True(t, bm.Get(12) == false)
	assert.True(t, bm.Get(74) == false)
	assert.True(t, bm.Get(2) == false)
	assert.True(t, bm.Get(13) == false)
	assert.True(t, bm.Get(75) == false)
	t.Log(bm.Cap())
}
