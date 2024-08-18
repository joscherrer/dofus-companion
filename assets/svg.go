package assets

import (
	_ "embed"

	"github.com/inkeliz/giosvg"
)

//go:embed arrow_up.svg
var arrowUp []byte
var ArrowUp = byteToIcon(arrowUp)

func byteToIcon(b []byte) *giosvg.Icon {
	v, err := giosvg.NewVector(b)
	if err != nil {
		panic(err)
	}
	return giosvg.NewIcon(v)
}

// var vector, err = giosvg.NewVector(arrowUp)

// if err != nil {
//     panic(err)
// }
//
// icon := giosvg.NewIcon(vector)
