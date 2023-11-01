package common

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

const initBitmapLen = 32
const defaultBitmapCap = 1 << 32

type Bitmap struct {
	arr      []byte
	ones     int
	capacity int
}

// NewBitmap capacity: bitmap max len in bits, default is 2^32 (about 0.5GiB memory)
func NewBitmap(capacity ...int) *Bitmap {
	bm := &Bitmap{arr: make([]byte, initBitmapLen)}
	if len(capacity) > 0 && capacity[0] > 0 {
		bm.capacity = capacity[0]
	} else {
		bm.capacity = defaultBitmapCap
	}
	return bm
}

func (b *Bitmap) Set(pos int) error {
	if pos < 0 || pos >= b.capacity {
		return errors.New("out of range")
	}
	if pos >= len(b.arr)<<3 {
		b.arr = append(b.arr, make([]byte, ((pos>>3)+1)-len(b.arr))...)
	}
	// &0x7 == %8
	if b.arr[pos>>3]&(1<<uint(pos&0x7)) == 0 {
		b.ones++
		b.arr[pos>>3] |= 1 << uint(pos&0x7)
	}
	return nil
}

func (b *Bitmap) Unset(pos int) {
	if pos < 0 || pos >= len(b.arr)<<3 {
		return
	}
	if b.arr[pos>>3]&(1<<uint(pos&0x7)) != 0 {
		b.ones--
		b.arr[pos>>3] &= ^(1 << uint(pos&0x7))
	}
}

func (b *Bitmap) Get(pos int) bool {
	if pos < 0 || pos >= len(b.arr)<<3 {
		return false
	}
	return b.arr[pos>>3]&(1<<uint(pos&0x7)) != 0
}

// GetFirstUnset get first unset bit
func (b *Bitmap) GetFirstUnset(start ...int) int {
	if len(start) == 0 {
		start = append(start, 0)
	}

	i := start[0] >> 3  // start[0] / 8
	j := start[0] & 0x7 // start[0] % 8

	for ; i < len(b.arr); i++ {
		if b.arr[i] == 0b11111111 {
			j = 0
			continue
		}

		for ; j < 8; j++ {
			if b.arr[i]&(1<<uint(j)) == 0 {
				return (i << 3) | j // i * 8 + j
			}
		}
		j = 0
	}

	pos := (i << 3) | j // i * 8 + j

	if pos >= b.capacity {
		return -1
	}
	return pos
}

// GetFirstSet get first set bit
func (b *Bitmap) GetFirstSet(start ...int) int {
	if len(start) == 0 {
		start = append(start, 0)
	}

	i := start[0] >> 3  // start[0] / 8
	j := start[0] & 0x7 // start[0] % 8

	for ; i < len(b.arr); i++ {
		if b.arr[i] == 0x00 {
			j = 0
			continue
		}

		for ; j < 8; j++ {
			if b.arr[i]&(1<<uint(j)) != 0 {
				return (i << 3) | j // i * 8 + j
			}
		}
		j = 0
	}

	pos := (i << 3) | j // i * 8 + j

	if pos >= b.capacity {
		return -1
	}
	return pos
}

func (b *Bitmap) Cap() int {
	return b.capacity
}

func (b *Bitmap) Ones() int {
	return b.ones
}

func (b *Bitmap) String() string {
	sb := strings.Builder{}
	sb.WriteString("Bitmap(")
	sb.WriteString("cap:")
	sb.WriteString(strconv.Itoa(b.Cap()))
	sb.WriteString(", ones:")
	sb.WriteString(strconv.Itoa(b.Ones()))
	sb.WriteString(") [")

	for i := 0; i < 32; i++ {
		val := "0"
		if b.Get(i) {
			val = "1"
		}
		sb.WriteString(val)
		sb.WriteString(" ")
	}
	sb.WriteString("...]")
	return sb.String()
}

// 匿名结构体，用于存储Bitmap的公开表示
type alias struct {
	Arr      []byte `json:"arr"`
	Ones     int    `json:"ones"`
	Capacity int    `json:"capacity"`
}

// MarshalJSON implements json.Marshaler interface.
func (b *Bitmap) MarshalJSON() ([]byte, error) {
	return json.Marshal(&alias{
		Arr:      b.arr,
		Ones:     b.ones,
		Capacity: b.capacity,
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (b *Bitmap) UnmarshalJSON(data []byte) error {
	a := &alias{}
	if err := json.Unmarshal(data, a); err != nil {
		return err
	}
	// 将公开表示的值复制到Bitmap对象
	b.arr = a.Arr
	b.ones = a.Ones
	b.capacity = a.Capacity
	return nil
}
