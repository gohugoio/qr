Basic QR encoder with direct PNG output.

This is a trimmed-down fork of [rsc.io/qr](https://github.com/rsc/qr), maintained for use in [Hugo](https://gohugo.io).

    go get github.com/gohugoio/qr

## Packages

* `qr`: the public API. `qr.Encode` returns a `*Code`, which implements `image.Image` and can be written directly as PNG with `Code.PNG`.
* `coding`: the low-level QR encoder (versions, levels, masks, bit layout).
* `gf256`: Galois field arithmetic used for Reed-Solomon error correction.

`coding` is tested against reference output recorded from the C libqrencode library, see `coding/testdata/libqrencode.golden`.
