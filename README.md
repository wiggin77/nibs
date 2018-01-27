[![Build Status](https://travis-ci.org/wiggin77/nibs.svg?branch=master)](https://travis-ci.org/wiggin77/nibs)

# nibs
Nibs is a Go package for reading streams of bytes in nibbles. Each nibble can be a number of bits in the range 1 bit to 64 bits inclusive.

See [godoc](https://godoc.org/github.com/wiggin77/nibs).

## Usage

```go
// provide an io.Reader
	b := []byte("this is a test")
	buf := bytes.NewReader(b)

// create instance of Nibs
nib := nibs.New(buf)

// read until io.EOF, 
  for {
    n, err := nib.Nibble8(4)
    if err != nil {
      break
    }
    fmt.Printf("nibbled 4 bits: %d", n)
  }

  // once EOF is reached, there will be bits left over
  // if nibble sizes do not divide evenly into the `NibbleXX` 
  // return type. Use `BitsRemaining` to determine how many.
  remaining, err := nib.BitsRemaining()
  if err == nil && remaining > 0 {
    n, err := nib.Nibble(remaining)
    if err == nil {
      fmt.Printf("nibbled %d remaining bits: %d", remaining, n)
    }
  }
}
```

