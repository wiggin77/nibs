package nibs

import (
	"errors"
	"io"
)

var (
	// ErrNibbleSize is the error used when an invalid nibble size is passed to a read method.
	ErrNibbleSize = errors.New("invalid nibble size (must be 1-64 bits inclusive)")

	// ErrNotEOF is the error used when trying to check how many bits are remaining but
	// not yet at end of file.
	ErrNotEOF = errors.New("not at EOF")
)

// Nibs reads a stream of bytes in nibbles of 1 bit to 64 bits.
type Nibs struct {
	reader   io.Reader
	curr     [8]byte
	usedCurr int // how many bytes in `curr` were read
	next     [8]byte
	usedNext int  // how many bytes in `next` were read
	pos      int  // bit position of next nibble (0-63)
	eof      bool // true if EOF is reached for `r` (`next` is empty)
}

// New returns a new Nibs which reads from the specified io.Reader.
func New(r io.Reader) *Nibs {
	return &Nibs{reader: r}
}

// BitsRemaining returns the number of bits that can be read once EOF is reached.
// Reading more than this number causes `Nibble` to return io.ErrUnexpectedEOF.
// If EOF is not yet reached then nibs.ErrNotEOF is returned.
// If error is nil and zero is returned then all the bits in the stream have been read.
func (n *Nibs) BitsRemaining() (int, error) {
	if !n.eof {
		return 0, ErrNotEOF
	}
	return n.remaining(), nil
}

// helper, likely inlined
func (n *Nibs) remaining() int {
	return (n.usedCurr * 8) - n.pos
}

// Nibble reads `bits` number of bits from the byte stream and returns the
// value as a uint64.
//
// `bits` must be in the range 1 to 64 inclusive, otherwise
// nibs.ErrNibbleSize is returned.
//
// io.EOF is returned when exactly all the bits in the stream have been read.
// io.ErrUnexpectedEOF is returned when trying to read more bits than are left
// in the stream.  A value of 0 is always returned for any non-nil error.
func (n *Nibs) Nibble(bits int) (uint64, error) {
	if bits < 1 || bits > 64 {
		return 0, ErrNibbleSize
	}
	// check if all bits already read or trying to read more bits than available
	if n.eof {
		remaining := n.remaining()
		if remaining == 0 {
			return 0, io.EOF
		}
		if bits > remaining {
			return 0, io.ErrUnexpectedEOF
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

func (n *Nibs) nextBit() (byte, error) {
	// check if we need to read more bytes
	if n.pos == n.usedCurr*8 {
		if n.eof {
			return 0, io.EOF
		}
		var err error
		// if first read, then read curr
		if n.usedCurr == 0 && n.usedNext == 0 {
			if n.usedCurr, err = n.reader.Read(n.curr[:]); err != nil {
				if err == io.EOF {
					n.eof = true
				} else {
					return 0, err
				}
			}
		} else {
			// not first read so promote next to curr
			n.curr = n.next
			n.usedCurr = n.usedNext
			n.pos = 0
		}
		// read next
		if n.usedNext, err = n.reader.Read(n.next[:]); err != nil {
			if err == io.EOF {
				n.eof = true
			} else {
				return 0, err
			}
		}
	}

	// get the correct byte based on pos
	b := n.curr[n.pos/8]
	// shift the bit we want to the rightmost
	b = b >> (8 - uint(n.pos%8) - 1)
	// increment pos to next position
	n.pos++
	// return 1 or 0
	return b & 1, nil
}
