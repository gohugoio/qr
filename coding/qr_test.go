// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coding

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gohugoio/qr/gf256"
)

var input = []Encoding{
	String("hello"),
	Num("1"),
	Num("12"),
	Num("123"),
	Alpha("AB"),
	Alpha("ABC"),
}

// gridHash returns the hash used in testdata/libqrencode.golden.
func gridHash(grid [][]byte) string {
	h := sha256.New()
	for _, row := range grid {
		h.Write(row)
	}
	return hex.EncodeToString(h.Sum(nil)[:16])
}

// planGrid renders an encoded plan in the golden grid alphabet.
func planGrid(p *Plan, c *Code) [][]byte {
	grid := make([][]byte, len(p.Pixel))
	for y, row := range p.Pixel {
		grid[y] = make([]byte, len(row))
		for x, pix := range row {
			var b byte
			switch pix.Role() {
			case Position:
				b = 'p'
			case Alignment:
				b = 'a'
			case Timing:
				b = 't'
			case Format:
				b = 'f'
			case PVersion:
				b = 'v'
			case Data, Check, Extra:
				b = 'd'
			default:
				b = 'u'
			}
			if c.Black(x, y) {
				b -= 'a' - 'A'
			}
			grid[y][x] = b
		}
	}
	return grid
}

// TestVersion checks every version, level and a few inputs against output
// recorded from libqrencode.
func TestVersion(t *testing.T) {
	f, err := os.Open("testdata/libqrencode.golden")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	levels := map[string]Level{"L": L, "M": M, "Q": Q, "H": H}
	n, bad := 0, 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 7 {
			t.Fatalf("malformed golden line %q", line)
		}
		v, _ := strconv.Atoi(fields[0])
		l, ok := levels[fields[1]]
		if !ok {
			t.Fatalf("bad level in %q", line)
		}
		var enc Encoding
		switch fields[2] {
		case "String":
			enc = String(fields[3])
		case "Alpha":
			enc = Alpha(fields[3])
		case "Num":
			enc = Num(fields[3])
		default:
			t.Fatalf("bad encoding in %q", line)
		}
		mask, _ := strconv.Atoi(fields[4])
		size, _ := strconv.Atoi(fields[5])

		n++
		p, err := NewPlan(Version(v), l, Mask(mask))
		if err != nil {
			t.Errorf("NewPlan(%d, %v, %d): %v", v, l, mask, err)
			bad++
			continue
		}
		if len(p.Pixel) != size {
			t.Errorf("%d/%v: NewPlan uses %dx%d, libqrencode uses %dx%d", v, l, len(p.Pixel), len(p.Pixel), size, size)
			bad++
			continue
		}
		c, err := p.Encode(enc)
		if err != nil {
			t.Errorf("Encode(%v): %v", enc, err)
			bad++
			continue
		}
		if got := gridHash(planGrid(p, c)); got != fields[6] {
			t.Errorf("%d/%v/%d %v: grid hash %s, want %s", v, l, mask, enc, got, fields[6])
			bad++
		}
		if bad >= 10 {
			t.Fatalf("stopping after %d bad cases", bad)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if want := 40 * 4 * len(input); n != want {
		t.Errorf("golden file has %d cases, want %d", n, want)
	}
}

func TestEncode(t *testing.T) {
	data := []byte{0x10, 0x20, 0x0c, 0x56, 0x61, 0x80, 0xec, 0x11, 0xec, 0x11, 0xec, 0x11, 0xec, 0x11, 0xec, 0x11}
	check := []byte{0xa5, 0x24, 0xd4, 0xc1, 0xed, 0x36, 0xc7, 0x87, 0x2c, 0x55}
	rs := gf256.NewRSEncoder(Field, len(check))
	out := make([]byte, len(check))
	rs.ECC(data, out)
	if !bytes.Equal(out, check) {
		t.Errorf("have %x want %x", out, check)
	}
}
