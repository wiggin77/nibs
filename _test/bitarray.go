package _test

import (
	"bytes"
	"fmt"
)

// BitArray is a simple array of bits.
// For testing only - not optimized for memory or speed.
type BitArray struct {
	buf []bool
}

// Add adds a single bit to the bit array.
func (ba *BitArray) Add(bit bool) {
	ba.buf = append(ba.buf, bit)
}

// AddVar adds bits from `bits`, starting from the left-most bit, up
// to `sigBits` number of bits.
// sigBits must be 0-64 inclusive.
func (ba *BitArray) AddVar(bits uint64, sigBits int) {
	if sigBits < 0 || sigBits > 64 {
		panic(fmt.Errorf("sigBits out of range; expected 0-64, got %d", sigBits))
	}

	for i := 0; i < sigBits; i++ {
		n := (bits >> uint(64-i)) & 1
		ba.Add(n != 0)
	}
}

// Add8 adds the 8 bits to the array, starting from the left-most bit.
func (ba *BitArray) Add8(bits byte) {
	n := uint64(bits)
	n = n << 56
	ba.AddVar(n, 8)
}

// Add16 adds the 16 bits to the array, starting from the left-most bit.
func (ba *BitArray) Add16(bits uint16) {
	n := uint64(bits)
	n = n << 48
	ba.AddVar(n, 16)
}

// Add32 adds the 32 bits to the array, starting from the left-most bit.
func (ba *BitArray) Add32(bits uint32) {
	n := uint64(bits)
	n = n << 32
	ba.AddVar(n, 32)
}

// Add64 adds the 64 bits to the array, starting from the left-most bit.
func (ba *BitArray) Add64(bits uint64) {
	ba.AddVar(bits, 64)
}

// AddSlice adds all the bits in the slice of bytes to the bit array,
// starting with the left-most bits at slice index 0.
func (ba *BitArray) AddSlice(bits []byte) {
	for _, b := range bits {
		ba.Add8(b)
	}
}

// Equals returns true only if `barr` contains the same number of
// elements as this, and each element has the same value.
func (ba *BitArray) Equals(barr *BitArray) bool {
	if len(ba.buf) != len(barr.buf) {
		return false
	}
	for i, b := range ba.buf {
		if barr.buf[i] != b {
			return false
		}
	}
	return true
}

// String returns a string representation of the bit array.
func (ba *BitArray) String() string {
	buf := bytes.Buffer{}
	for _, b := range ba.buf {
		if b {
			buf.WriteString("1")
		} else {
			buf.WriteString("0")
		}
	}
	return buf.String()
}
