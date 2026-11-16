// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qr

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestPNG(t *testing.T) {
	for _, scale := range []int{1, 2, 3, 8} {
		c, err := Encode("hello, world", Options{Scale: scale})
		if err != nil {
			t.Fatal(err)
		}
		pngdat := c.PNG()
		m, err := png.Decode(bytes.NewBuffer(pngdat))
		if err != nil {
			t.Fatal(err)
		}
		gm := m.(*image.Gray)

		siz := c.Size
		nbad := 0
		for y := 0; y < scale*(8+siz); y++ {
			for x := 0; x < scale*(8+siz); x++ {
				v := byte(255)
				if c.Black(x/scale-4, y/scale-4) {
					v = 0
				}
				if gv := gm.At(x, y).(color.Gray).Y; gv != v {
					t.Errorf("scale %d: %d,%d = %d, want %d", scale, x, y, gv, v)
					if nbad++; nbad >= 20 {
						t.Fatalf("too many bad pixels")
					}
				}
			}
		}
	}
}

// TestPNGDefaultUnchanged checks that the default black on white output is
// byte for byte what the code produced before colors were supported.
func TestPNGDefaultUnchanged(t *testing.T) {
	want := []struct {
		level Level
		scale int
		hash  string
	}{
		{L, 2, "4a34ff4aab4ab8c1cefdde74e4d08416"},
		{L, 4, "cfcc5ebb11b1179950a10d0f82723dce"},
		{L, 8, "d5f610185a90a50ad916e7d1a4831789"},
		{M, 2, "083c2dbd57c9806c7809573447799813"},
		{M, 4, "7c812356478f6e94bb414470817fcf5d"},
		{M, 8, "70f6b0bb2865d6b3efb2992ee2b61291"},
		{Q, 2, "b32da362061393e542908f3049a404cb"},
		{Q, 4, "9c4741cb04d6ffcc79ac4ba9a9b91874"},
		{Q, 8, "ddf5241c73b184aad5430590de604de6"},
		{H, 2, "1b5b928800fba38f06e166d5408e0741"},
		{H, 4, "9b597f190cbcc3afe07ca532ef35620c"},
		{H, 8, "6e386604aeb457ff1933ce98ea999376"},
	}
	for _, w := range want {
		for _, opts := range []Options{
			{Level: w.level, Scale: w.scale},
			// Explicit black and white must give the same output.
			{Level: w.level, Scale: w.scale, FgColor: color.Black, BgColor: color.White},
		} {
			c, err := Encode("https://gohugo.io/", opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(c.PNG()))[:32]; got != w.hash {
				t.Errorf("%v/%d: PNG hash %s, want %s", w.level, w.scale, got, w.hash)
			}
		}
	}
}

func TestPNGColors(t *testing.T) {
	nrgba := func(c color.Color) color.NRGBA {
		return color.NRGBAModel.Convert(c).(color.NRGBA)
	}
	for _, tc := range []struct {
		name   string
		fg, bg color.Color
	}{
		{"transparent background", nil, color.Transparent},
		{"colors", color.RGBA{0x10, 0x20, 0x80, 0xFF}, color.NRGBA{0xFF, 0xEE, 0xDD, 0xFF}},
		{"semi-transparent", color.NRGBA{0x80, 0, 0, 0xC0}, color.NRGBA{0xFF, 0xFF, 0xFF, 0x40}},
		{"black on transparent white", color.Black, color.NRGBA{0xFF, 0xFF, 0xFF, 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := Encode("hello, world", Options{Level: M, Scale: 3, FgColor: tc.fg, BgColor: tc.bg})
			if err != nil {
				t.Fatal(err)
			}
			fg, bg := color.NRGBA{A: 0xFF}, color.NRGBA{0xFF, 0xFF, 0xFF, 0xFF}
			if tc.fg != nil {
				fg = nrgba(tc.fg)
			}
			if tc.bg != nil {
				bg = nrgba(tc.bg)
			}

			m, err := png.Decode(bytes.NewReader(c.PNG()))
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := m.(*image.Paletted); !ok {
				t.Fatalf("decoded %T, want *image.Paletted", m)
			}
			img := c.Image()
			if m.Bounds() != img.Bounds() {
				t.Fatalf("bounds %v, want %v", m.Bounds(), img.Bounds())
			}

			nbad := 0
			b := m.Bounds()
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					want := bg
					if c.Black(x/c.Scale-4, y/c.Scale-4) {
						want = fg
					}
					if got := nrgba(m.At(x, y)); got != want {
						t.Errorf("PNG %d,%d = %v, want %v", x, y, got, want)
						nbad++
					}
					if got := nrgba(img.At(x, y)); got != want {
						t.Errorf("Image %d,%d = %v, want %v", x, y, got, want)
						nbad++
					}
					if nbad >= 20 {
						t.Fatalf("too many bad pixels")
					}
				}
			}
		})
	}
}

// TestImage checks that Image draws the same picture as PNG.
func TestImage(t *testing.T) {
	c, err := Encode("hello, world", Options{Scale: 4})
	if err != nil {
		t.Fatal(err)
	}
	m, err := png.Decode(bytes.NewReader(c.PNG()))
	if err != nil {
		t.Fatal(err)
	}
	img := c.Image()
	if m.Bounds() != img.Bounds() {
		t.Fatalf("bounds %v, want %v", m.Bounds(), img.Bounds())
	}
	b := m.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if got, want := color.GrayModel.Convert(img.At(x, y)), m.At(x, y); got != want {
				t.Fatalf("Image %d,%d = %v, want %v", x, y, got, want)
			}
		}
	}
}

func TestEncodeOptions(t *testing.T) {
	c, err := Encode("x", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if c.Scale != DefaultScale {
		t.Errorf("Scale = %d, want %d", c.Scale, DefaultScale)
	}
	for _, opts := range []Options{
		{Scale: -1},
		{Level: H + 1},
		{Level: -1},
	} {
		if _, err := Encode("x", opts); err == nil {
			t.Errorf("Encode with %+v: expected error", opts)
		}
	}
}

func BenchmarkPNG(b *testing.B) {
	c, err := Encode("0123456789012345678901234567890123456789", Options{})
	if err != nil {
		panic(err)
	}
	var bytes []byte
	for i := 0; i < b.N; i++ {
		bytes = c.PNG()
	}
	b.SetBytes(int64(len(bytes)))
}

func BenchmarkImagePNG(b *testing.B) {
	c, err := Encode("0123456789012345678901234567890123456789", Options{})
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		buf.Reset()
		png.Encode(&buf, c.Image())
	}
	b.SetBytes(int64(buf.Len()))
}
