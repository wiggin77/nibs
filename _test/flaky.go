package _test

import (
	"errors"
	"io"
)

// ErrFlaky is error returned when simulating flaky IO.
var ErrFlaky = errors.New("flaky test")

// FlakyReader implements io.Reader but only reads `count` bytes
// before
type FlakyReader struct {
	reader io.Reader
	count  int
}

// NewFlakyReader returns a FlakyReader with the specified number of
// bytes read before ErrFlaky is returned.
func NewFlakyReader(r io.Reader, goodBytes int) *FlakyReader {
	return &FlakyReader{reader: r, count: goodBytes}
}

// Read reads from the data stream, returning ErrFlaky once the
// allowed number of good bytes is reached.
func (fr *FlakyReader) Read(p []byte) (int, error) {
	if fr.count <= 0 {
		return 0, ErrFlaky
	}

	// read less bytes than original slice would allow if
	// good byte count is less than buffer size
	bufSize := len(p)
	if bufSize > fr.count {
		p = p[:bufSize-fr.count-1]
	}

	n, err := fr.reader.Read(p)
	fr.count -= n
	if fr.count <= 0 {
		err = ErrFlaky
	}
	return n, err
}
