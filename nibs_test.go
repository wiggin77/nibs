package nibs_test

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"io"
	"testing"

	. "github.com/wiggin77/nibs/_test"

	"github.com/wiggin77/nibs"
)

func TestNew(t *testing.T) {
	var num uint64 = 123456789
	b := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(b, num)
	buf := bytes.NewReader(b)

	nib := nibs.New(buf)

	var result = make([]byte, binary.MaxVarintLen64)

	for i := 0; i < binary.MaxVarintLen64; i++ {
		n, err := nib.Nibble(8)
		if err != nil {
			t.Errorf("error reading byte %d: %v", i, err)
			return
		}
		result[i] = byte(n)
	}

	// should be zero bits remaining
	r, err := nib.BitsRemaining()
	if err != nil || r != 0 {
		t.Errorf("wrong bits remaining, expected 0, got %d and error `%v`", r, err)
	}

	// next read should be EOF
	_, err = nib.Nibble(8)
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}

	// result should match
	res64 := binary.LittleEndian.Uint64(result)
	if res64 != num {
		t.Errorf("result doesn't match, expected %d, got %d and error `%v`", num, res64, err)
	}
}

func TestErrNibbleSize(t *testing.T) {
	b := []byte("this is a test")
	buf := bytes.NewReader(b)
	nib := nibs.New(buf)

	if _, err := nib.Nibble(-1); err != nibs.ErrNibbleSize {
		t.Errorf("expected `nibs.ErrNibbleSize`, got %v", err)
	}

	if _, err := nib.Nibble(0); err != nibs.ErrNibbleSize {
		t.Errorf("expected `nibs.ErrNibbleSize`, got %v", err)
	}

	if _, err := nib.Nibble(65); err != nibs.ErrNibbleSize {
		t.Errorf("expected `nibs.ErrNibbleSize`, got %v", err)
	}

	// should be no error
	if _, err := nib.Nibble(2); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEOF(t *testing.T) {
	b := []byte{1, 2, 4, 8, 16, 32, 64, 128}
	buf := bytes.NewReader(b)
	nib := nibs.New(buf)

	// read 6 bytes
	if _, err := nib.Nibble(48); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// should be 2 bytes left
	if n, err := nib.BitsRemaining(); err != nil || n != 16 {
		t.Errorf("expected %d bits remaining, got %d and error `%v`", 16, n, err)
	}

	// reading 4 bytes should error
	if _, err := nib.Nibble(32); err != io.EOF {
		t.Errorf("expected error `io.EOF`, got `%v`", err)
	}

	// reading 2 bytes should be ok
	if _, err := nib.Nibble(16); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// reading reading any more should be EOF
	if _, err := nib.Nibble(48); err != io.EOF {
		t.Errorf("expected error `io.EOF`, got `%v`", err)
	}
}

func TestTiny(t *testing.T) {
	b := []byte{0xFF}
	buf := bytes.NewReader(b)
	nib := nibs.New(buf)

	// nibble 8 bits, all should be 1
	for i := 0; i < 8; i++ {
		bit, err := nib.Nibble(1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if bit != 1 {
			t.Errorf("expected 1, got %d", bit)
		}
	}

	// reading any more should be EOF
	if _, err := nib.Nibble(1); err != io.EOF {
		t.Errorf("expected error `io.EOF`, got `%v`", err)
	}

	// should be 0 bits left
	if n, err := nib.BitsRemaining(); err != nil || n != 0 {
		t.Errorf("expected %d bits remaining, got %d and error `%v`", 0, n, err)
	}
}

func TestMany(t *testing.T) {
	sizes := []int{0, 1, 2, 10, 64, 128, 1024, 2048, 10000, 10 * 1024 * 1000}
	for _, size := range sizes {
		trial(size, t)
	}
}

func trial(size int, t *testing.T) {
	bufIn := make([]byte, size)
	if _, err := rand.Read(bufIn); err != nil {
		panic(err)
	}
	nib := nibs.New(bytes.NewReader(bufIn))

	// nibble 4 bits at at time (hey, an actual nibble)
	bufOut := make([]byte, size)
	var b byte
	for i := 0; i < size; i++ {
		n, err := nib.Nibble(4)
		if err != nil {
			t.Errorf("unexpected error in first nibble for size %d, byte %d: %v", size, i, err)
		}
		b = (byte(n) << 4) & 0xF0

		n, err = nib.Nibble(4)
		if err != nil {
			t.Errorf("unexpected error in second nibble for size %d, byte %d: %v", size, i, err)
		}
		b = b | (byte(n) & 0xFF)
		bufOut[i] = b
	}

	// reading any more should be EOF
	if _, err := nib.Nibble(1); err != io.EOF {
		t.Errorf("expected error `io.EOF`, got `%v` for size %d", err, size)
	}

	// should be 0 bits left
	if n, err := nib.BitsRemaining(); err != nil || n != 0 {
		t.Errorf("expected %d bits remaining, got %d and error `%v` for size %d", 0, n, err, size)
	}

	// compare results
	if !bytes.Equal(bufIn, bufOut) {
		t.Errorf("bufIn != bufOut for size %d", size)
	}
}

func TestNibble8(t *testing.T) {
	const size = 256
	bufIn := make([]byte, size)
	if _, err := rand.Read(bufIn); err != nil {
		panic(err)
	}
	nib := nibs.New(bytes.NewReader(bufIn))

	bufOut := make([]byte, size)
	for i := 0; i < size; i++ {
		b, err := nib.Nibble8(8)
		if err != nil {
			t.Errorf("unexpected error `%v`", err)
			break
		}
		bufOut[i] = b
	}
	if !bytes.Equal(bufIn, bufOut) {
		t.Error("bufIn != bufOut")
	}
}

func TestNibble16(t *testing.T) {
	const size = 256
	bufIn := make([]byte, size)
	if _, err := rand.Read(bufIn); err != nil {
		panic(err)
	}
	nib := nibs.New(bytes.NewReader(bufIn))

	bufOut := make([]byte, size)
	for i := 0; i < size; i++ {
		b, err := nib.Nibble16(8)
		if err != nil {
			t.Errorf("unexpected error `%v`", err)
			break
		}
		bufOut[i] = byte(b)
	}
	if !bytes.Equal(bufIn, bufOut) {
		t.Error("bufIn != bufOut")
	}
}

func TestNibble32(t *testing.T) {
	const size = 256
	bufIn := make([]byte, size)
	if _, err := rand.Read(bufIn); err != nil {
		panic(err)
	}
	nib := nibs.New(bytes.NewReader(bufIn))

	bufOut := make([]byte, size)
	for i := 0; i < size; i++ {
		b, err := nib.Nibble32(8)
		if err != nil {
			t.Errorf("unexpected error `%v`", err)
			break
		}
		bufOut[i] = byte(b)
	}
	if !bytes.Equal(bufIn, bufOut) {
		t.Error("bufIn != bufOut")
	}
}

// test using every possible nibble size
func TestNibbleSizes(t *testing.T) {
	const size = 768
	for nibbleSize := 1; nibbleSize <= 64; nibbleSize++ {
		bufIn := make([]byte, size)
		if _, err := rand.Read(bufIn); err != nil {
			panic(err)
		}
		nib := nibs.New(bytes.NewReader(bufIn))

		baIn := &BitArray{}
		baIn.AddSlice(bufIn)

		baOut := &BitArray{}
		for {
			n, err := nib.Nibble(nibbleSize)
			if err != nil {
				break
			}
			baOut.AddVar(n, nibbleSize)
		}
		remaining, err := nib.BitsRemaining()
		if err != nil {
			t.Errorf("bits remaining should be known here. err=%v", err)
			break
		}
		if remaining > 0 {
			n, err := nib.Nibble(remaining)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				break
			}
			baOut.AddVar(n, remaining)
		}

		// now check result
		if !baOut.Equals(baOut) {
			t.Errorf("baIn != baOut for nibbleSize %d", nibbleSize)
		}
	}
}

func TestBitsRemaining(t *testing.T) {
	bufIn := make([]byte, 10*1000*1024)
	nib := nibs.New(bytes.NewReader(bufIn))

	if _, err := nib.BitsRemaining(); err != nibs.ErrUnknown {
		t.Error("expected ErrUnknown, got ", err)
	}
}

func TestNibbleSizeErrors(t *testing.T) {
	bufIn := make([]byte, 256)
	nib := nibs.New(bytes.NewReader(bufIn))

	if _, err := nib.Nibble32(64); err != nibs.ErrNibbleSize {
		t.Error("expected ErrNibbleSize for Nibble32, got ", err)
	}

	if _, err := nib.Nibble16(32); err != nibs.ErrNibbleSize {
		t.Error("expected ErrNibbleSize for Nibble16, got ", err)
	}

	if _, err := nib.Nibble8(16); err != nibs.ErrNibbleSize {
		t.Error("expected ErrNibbleSize for Nibble8, got ", err)
	}
}

// test nibbling from a flaky reader
func TestIOError(t *testing.T) {
	const size = 2 * 1024 * 1000
	const goodBytes = size / 2
	bufIn := make([]byte, size)
	if _, err := rand.Read(bufIn); err != nil {
		panic(err)
	}
	r := NewFlakyReader(bytes.NewReader(bufIn), goodBytes)
	nib := nibs.New(r)

	bufOut := make([]byte, len(bufIn))
	for i := 0; i < len(bufOut); i++ {
		b, err := nib.Nibble(8)

		if err == nil && i+1 > goodBytes {
			t.Errorf("expected a flaky error at byte %d", i)
		}

		if err != nil && i+1 <= goodBytes {
			t.Errorf("unexpected error at byte %d: %v", i, err)
		} else {
			bufOut[i] = byte(b)
		}
	}

	// compare should be ok up to `goodBytes`
	if !bytes.Equal(bufIn[:goodBytes], bufOut[:goodBytes]) {
		t.Error("goodBytes compare failed")
	}

	// overall compare should fail
	if bytes.Equal(bufIn, bufOut) {
		t.Error("full array compare should fail")
	}
}
