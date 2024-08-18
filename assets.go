package main

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
)

//go:embed assets/dofus_icon_128.png
var dofusIcon128 []byte
var DofusIcon128, _, _ = image.Decode(bytes.NewReader(dofusIcon128))

// var DofusIcon128 = widget.Icon{src: dofusIcon128}
