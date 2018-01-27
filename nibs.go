package nibs

import (
	"errors"
	"fmt"
	"io"
)

const (
	bufSize       = 64
	readThreshold = bufSize - 16 // point at which another read is needed
)

var (
	// ErrNibbleSize is the error used when an invalid nibble size is passed to a read method.
	ErrNibbleSize = errors.New("invalid nibble size")

	// ErrUnknown is the error used when requesting the number of bits left until EOF
	// and the answer is not yet known because EOF is not reached.
	ErrUnknown = errors.New("not at EOF")
)

// Nibs reads a stream of bytes in nibbles of 1 bit to 64 bits.
type Nibs struct {
	reader io.Reader
	buf    [bufSize]byte
	used   int   // number of bytes read into buf
	pos    int   // bit position of next nibble within buf (0-512)
	err    error // error after last used byte in curr
}

// New returns a new Nibs which reads from the specified io.Reader.
func New(r io.Reader) *Nibs {
	return &Nibs{reader: r}
}

// BitsRemaining returns the number of bits that are remaining to be read, if known.
// If not known, meaning EOF is not yet reached internally and no other IO errors have occured,
// then ErrUnknown is returned.
// If known, reading more than this number causes `Nibble` to return io.EOF.
// If error is nil and zero is returned then all the bits in the stream have been read.
func (n *Nibs) BitsRemaining() (int, error) {
	if n.err == nil {
		return 0, ErrUnknown
	}
	return n.remaining(), nil
}

// helper, likely inlined
func (n *Nibs) remaining() int {
	return (n.used * 8) - n.pos
}

// Nibble reads `bits` number of bits from the byte stream and returns the
// value as a uint64.
//
// `bits` must be in the range 1 to 64 inclusive, otherwise
// nibs.ErrNibbleSize is returned.
//
// io.EOF is returned on subsequent call when exactly all the bits in the
// stream have been read. io.EOF is returned immediately when trying to read
// more bits than are left in the stream.  Use `BitsRemaining` to see how many
// bits are left over after io.EOF.
//
// A value of 0 is always returned for any non-nil error.
func (n *Nibs) Nibble(bits int) (uint64, error) {
	if bits < 1 || bits > 64 {
		return 0, ErrNibbleSize
	}
	// check if all bits already read or trying to read more bits than available
	if n.err != nil {
		remaining := n.remaining()
		if remaining == 0 {
			return 0, n.err
		}
		if bits > remaining {
			return 0, io.EOF
		}
	}

	var ret uint64
	for i := 0; i < bits; i++ {
		bit, err := n.nextBit()
		if err != nil {
			return 0, err
		}
		ret = ret << 1
		bit64 := uint64(bit)
		ret = ret | uint64(bit64)
	}
	return ret, nil
}

// Nibble8  reads `bits` number of bits from the byte stream and returns the
// value as a uint8.
//
// `bits` must be in the range 1 to 8 inclusive, otherwise
// nibs.ErrNibbleSize is returned.
//
// See `Nibble` method for details.
func (n *Nibs) Nibble8(bits int) (uint8, error) {
	if bits < 1 || bits > 8 {
		return 0, ErrNibbleSize
	}
	val, err := n.Nibble(bits)
	return uint8(val), err
}

// Nibble16  reads `bits` number of bits from the byte stream and returns the
// value as a uint16.
//
// `bits` must be in the range 1 to 16 inclusive, otherwise
// nibs.ErrNibbleSize is returned.
//
// See `Nibble` method for details.
func (n *Nibs) Nibble16(bits int) (uint16, error) {
	if bits < 1 || bits > 16 {
		return 0, ErrNibbleSize
	}
	val, err := n.Nibble(bits)
	return uint16(val), err
}

// Nibble32  reads `bits` number of bits from the byte stream and returns the
// value as a uint32.
//
// `bits` must be in the range 1 to 32 inclusive, otherwise
// nibs.ErrNibbleSize is returned.
//
// See `Nibble` method for details.
func (n *Nibs) Nibble32(bits int) (uint32, error) {
	if bits < 1 || bits > 32 {
		return 0, ErrNibbleSize
	}
	val, err := n.Nibble(bits)
	return uint32(val), err
}

func (n *Nibs) nextBit() (byte, error) {
	var bpos = n.pos / 8             // byte index
	var bposOffset = uint(n.pos % 8) // bit offset within byte

	// check if we need to read more bytes.
	if bposOffset == 0 && (bpos == readThreshold || bpos == n.used) {
		if n.err == nil {
			// prep for read
			if bpos > 0 {
				c := copy(n.buf[:], n.buf[bpos:n.used])
				n.used = c
				n.pos = 0
				bpos = 0
			}
			// read more
			rbuf := n.buf[n.used:]
			c, err := n.reader.Read(rbuf)
			n.used += c
			if err != nil {
				n.err = err
			} else if c < len(rbuf) {
				// we got less than expected and no error; try to force the EOF
				rbuf = n.buf[n.used:]
				c, err = n.reader.Read(rbuf)
				n.used += c
				if err != nil {
					n.err = err
				}
			}
		}
	}

	// check if we exhausted the available bits.
	if bpos == n.used {
		if n.err != nil {
			return 0, n.err
		}
		return 0, fmt.Errorf("BUG: read past avail bytes, but no EOF or error. used=%d, pos=%d", n.used, n.pos)
	}

	// get the correct byte based on pos
	b := n.buf[bpos]
	// shift the bit we want to the rightmost
	b = b >> (8 - bposOffset - 1)
	// increment pos to next bit position
	n.pos++
	// return 1 or 0
	return b & 1, nil
}
