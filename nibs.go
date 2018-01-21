package nibs

import (
	"bufio"
	"errors"
	"io"
)

var (
	// ErrNibbleSize is error used when an invalid nibble size is passed to a read method.
	ErrNibbleSize = errors.New("invalid nibble size (must be 1-64 bits inclusive)")
)

// Nibs reads a stream of bytes in nibbles of 1 bit to 64 bits.
type Nibs struct {
	r   io.ByteReader
	b   byte
	pos byte
}

// New returns a new Nibs which reads from the specified io.Reader.
func New(r io.Reader) *Nibs {
	return &Nibs{r: bufio.NewReader(r)}
}

// Nibble reads `bits` number of bits from the byte stream and returns the
// value as a int64.
//
// `bits` must be in the range 1 to 64 inclusive, otherwise
// nibs.ErrNibbleSize is returned. A value of 0 is always returned for any
// non-nil error.
//
//
func (n *Nibs) Nibble(int bits) (int64, error) {

}

// Remainder returns the value of any bits leftover in the current byte.
// Use this once io.EOF is reached to check if any bits were left over.
func (n *Nibs) Remainder() byte {

}
