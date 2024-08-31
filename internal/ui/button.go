package ui

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/joscherrer/dofus-manager/internal/f32color"
)

type ImgBtnStyle struct {
	Background  color.NRGBA
	Color       color.NRGBA
	Image       *image.Image
	Size        unit.Dp
	Inset       layout.Inset
	Button      *Clickable
	Description string
}

type IconBtnStyle struct {
	Background color.NRGBA
	// Color is the icon color.
	Color color.NRGBA
	Icon  *widget.Icon
	// Size is the icon size.
	Size        unit.Dp
	Inset       layout.Inset
	Button      *Clickable
	Description string
}

type AssetBtnStyle struct {
	Background color.NRGBA
	// Color is the icon color.
	Color color.NRGBA
	Asset *Asset
	// Size is the icon size.
	Size        unit.Dp
	Inset       layout.Inset
	Button      *Clickable
	Description string
}

type Asset = struct {
	ViewBox struct{ Min, Max f32.Point }
	Call    op.CallOp
}

// StlImageButtonLayoutStyle ImageButtonStyle
// BtnLayoutStyle            material.ButtonLayoutStyle
// BtnStyle                  material.ButtonStyle
// IconBtnStyle material.IconButtonStyle

type BtnLayoutStyle struct {
	Background   color.NRGBA
	CornerRadius unit.Dp
	Button       *Clickable
	KeepState    bool
	Inset        layout.Inset
}

type BtnStyle struct {
	Text string
	// Color is the text color.
	Color        color.NRGBA
	Font         font.Font
	TextSize     unit.Sp
	Background   color.NRGBA
	CornerRadius unit.Dp
	Inset        layout.Inset
	Button       *Clickable
	shaper       *text.Shaper
	KeepState    bool
}

type SvgBtnStyle struct{}

func (b BtnLayoutStyle) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	min := gtx.Constraints.Min
	btn := b.Button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		return layout.Background{}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				rr := gtx.Dp(b.CornerRadius)
				defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, rr).Push(gtx.Ops).Pop()
				background := b.Background
				switch {
				case !gtx.Enabled():
					background = f32color.Disabled(b.Background)
				case b.Button.Hovered() || (b.KeepState && gtx.Focused(b.Button)) || b.Button.Active:
					background = f32color.Hovered(b.Background)
				}
				paint.Fill(gtx.Ops, background)
				for _, c := range b.Button.History() {
					drawInk(gtx, c)
				}
				return layout.Dimensions{Size: gtx.Constraints.Min}
			},
			func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = min
				return layout.Center.Layout(gtx, w)
			},
		)
	})
	return b.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return btn })
}

func (b BtnStyle) Layout(gtx layout.Context) layout.Dimensions {
	return BtnLayoutStyle{
		Background:   b.Background,
		CornerRadius: unit.Dp(4),
		Button:       b.Button,
		KeepState:    b.KeepState,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return b.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			colMacro := op.Record(gtx.Ops)
			paint.ColorOp{Color: b.Color}.Add(gtx.Ops)
			return widget.Label{Alignment: text.Middle}.Layout(gtx, b.shaper, b.Font, b.TextSize, b.Text, colMacro.Stop())
		})
	})
}

func (b IconBtnStyle) Layout(gtx layout.Context) layout.Dimensions {
	return BtnLayoutStyle{
		Background:   b.Background,
		CornerRadius: unit.Dp(4),
		Button:       b.Button,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return b.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			size := gtx.Dp(b.Size)
			gtx.Constraints.Max = image.Point{X: size, Y: size}
			return b.Icon.Layout(gtx, b.Color)
		})
	})
}

func (b AssetBtnStyle) Layout(gtx layout.Context) layout.Dimensions {
	return BtnLayoutStyle{
		Background:   b.Background,
		CornerRadius: unit.Dp(4),
		Button:       b.Button,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return b.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			size := gtx.Dp(b.Size)
			gtx.Constraints.Max = image.Point{X: size, Y: size}
			b.Asset.Call.Add(gtx.Ops)
			return layout.Dimensions{Size: image.Point{X: size, Y: size}}
		})
	})
}

func Btn(th *material.Theme, button *Clickable, text string, keepState bool) BtnStyle {
	b := BtnStyle{
		Text:         text,
		Color:        th.Palette.ContrastFg,
		CornerRadius: 4,
		Background:   th.Palette.ContrastBg,
		TextSize:     th.TextSize * 14.0 / 16.0,
		Inset: layout.Inset{
			Top: 10, Bottom: 10,
			Left: 12, Right: 12,
		},
		Button:    button,
		shaper:    th.Shaper,
		KeepState: keepState,
	}
	b.Font.Typeface = th.Face
	return b
}

func IconBtn(th *material.Theme, button *Clickable, icon *widget.Icon, description string) IconBtnStyle {
	return IconBtnStyle{
		Background:  th.Palette.ContrastBg,
		Color:       th.Palette.ContrastFg,
		Icon:        icon,
		Size:        22,
		Inset:       layout.UniformInset(8),
		Button:      button,
		Description: description,
	}
}

func AssetBtn(th *material.Theme, button *Clickable, asset *Asset, description string) AssetBtnStyle {
	return AssetBtnStyle{
		Background:  th.Palette.ContrastBg,
		Color:       th.Palette.ContrastFg,
		Asset:       asset,
		Size:        14,
		Inset:       layout.UniformInset(12),
		Button:      button,
		Description: description,
	}
}

func NewBtn(theme *material.Theme, wc *Clickable, text string, keepState bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return Btn(theme, wc, text, keepState).Layout(gtx)
	}
}

// func NewButton(theme *material.Theme, wc *Clickable, text string) layout.Widget {
// 	return func(gtx layout.Context) layout.Dimensions {
// 		return material.Button(theme, wc, text).Layout(gtx)
// 	}
// }

func NewIconBtn(theme *material.Theme, wc *Clickable, icon *widget.Icon, description string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return IconBtn(theme, wc, icon, description).Layout(gtx)
	}
}

func NewAssetBtn(theme *material.Theme, wc *Clickable, asset *Asset, description string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return AssetBtn(theme, wc, asset, description).Layout(gtx)
	}
}

// func StlIconButton(th *material.Theme, button *Clickable, icon *widget.Icon, description string) IconBtnStyle {
// 	return IconBtnStyle{
// 		Background:  th.Palette.ContrastBg,
// 		Color:       th.Palette.ContrastFg,
// 		Icon:        icon,
// 		Size:        22,
// 		Inset:       layout.UniformInset(8),
// 		Button:      button,
// 		Description: description,
// 	}
// }

// func NewImageButton(theme *material.Theme, wc *Clickable, img *image.Image, description string) layout.Widget {
// 	return func(gtx layout.Context) layout.Dimensions {
// 		return ImageButton(theme, wc, img, description).Layout(gtx)
// 	}
// }

// func (b ImageButtonStyle) Layout(gtx layout.Context) layout.Dimensions {
// 	return material.ButtonLayoutStyle{
// 		Background:   b.Background,
// 		CornerRadius: unit.Dp(4),
// 		Button:       b.Button,
// 	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
// 		return b.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
// 			size := gtx.Dp(b.Size)
// 			gtx.Constraints.Max = image.Point{X: size, Y: size}
// 			img := widget.Image{Src: paint.NewImageOp(*b.Image), Fit: widget.Cover}
// 			return img.Layout(gtx)
// 		})
// 	})
// }

// func (b StlButtonStyle) Layout(gtx layout.Context) layout.Dimensions {
// 	return BtnLayoutStyle{
// 		Background:   b.Background,
// 		CornerRadius: b.CornerRadius,
// 		Button:       b.Button,
// 	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
// 		return b.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
// 			colMacro := op.Record(gtx.Ops)
// 			paint.ColorOp{Color: b.Color}.Add(gtx.Ops)
// 			return widget.Label{Alignment: text.Middle}.Layout(gtx, b.shaper, b.Font, b.TextSize, b.Text, colMacro.Stop())
// 		})
// 	})
// }

// func ImageButton(th *material.Theme, button *Clickable, image *image.Image, description string) ImageButtonStyle {
// 	return ImageButtonStyle{
// 		Background:  th.Palette.ContrastBg,
// 		Color:       th.Palette.ContrastFg,
// 		Image:       image,
// 		Size:        14,
// 		Inset:       layout.UniformInset(12),
// 		Button:      button,
// 		Description: description,
// 	}
// }

func drawInk(gtx layout.Context, c widget.Press) {
	// duration is the number of seconds for the
	// completed animation: expand while fading in, then
	// out.
	const (
		expandDuration = float32(0.5)
		fadeDuration   = float32(0.9)
	)

	now := gtx.Now

	t := float32(now.Sub(c.Start).Seconds())

	end := c.End
	if end.IsZero() {
		// If the press hasn't ended, don't fade-out.
		end = now
	}

	endt := float32(end.Sub(c.Start).Seconds())

	// Compute the fade-in/out position in [0;1].
	var alphat float32
	{
		var haste float32
		if c.Cancelled {
			// If the press was cancelled before the inkwell
			// was fully faded in, fast forward the animation
			// to match the fade-out.
			if h := 0.5 - endt/fadeDuration; h > 0 {
				haste = h
			}
		}
		// Fade in.
		half1 := t/fadeDuration + haste
		if half1 > 0.5 {
			half1 = 0.5
		}

		// Fade out.
		half2 := float32(now.Sub(end).Seconds())
		half2 /= fadeDuration
		half2 += haste
		if half2 > 0.5 {
			// Too old.
			return
		}

		alphat = half1 + half2
	}

	// Compute the expand position in [0;1].
	sizet := t
	if c.Cancelled {
		// Freeze expansion of cancelled presses.
		sizet = endt
	}
	sizet /= expandDuration

	// Animate only ended presses, and presses that are fading in.
	if !c.End.IsZero() || sizet <= 1.0 {
		gtx.Execute(op.InvalidateCmd{})
	}

	if sizet > 1.0 {
		sizet = 1.0
	}

	if alphat > .5 {
		// Start fadeout after half the animation.
		alphat = 1.0 - alphat
	}
	// Twice the speed to attain fully faded in at 0.5.
	t2 := alphat * 2
	// Beziér ease-in curve.
	alphaBezier := t2 * t2 * (3.0 - 2.0*t2)
	sizeBezier := sizet * sizet * (3.0 - 2.0*sizet)
	size := gtx.Constraints.Min.X
	if h := gtx.Constraints.Min.Y; h > size {
		size = h
	}
	// Cover the entire constraints min rectangle and
	// apply curve values to size and color.
	size = int(float32(size) * 2 * float32(math.Sqrt(2)) * sizeBezier)
	alpha := 0.7 * alphaBezier
	const col = 0.8
	ba, bc := byte(alpha*0xff), byte(col*0xff)
	rgba := f32color.MulAlpha(color.NRGBA{A: 0xff, R: bc, G: bc, B: bc}, ba)
	ink := paint.ColorOp{Color: rgba}
	ink.Add(gtx.Ops)
	rr := size / 2
	defer op.Offset(c.Position.Add(image.Point{
		X: -rr,
		Y: -rr,
	})).Push(gtx.Ops).Pop()
	defer clip.UniformRRect(image.Rectangle{Max: image.Pt(size, size)}, rr).Push(gtx.Ops).Pop()
	paint.PaintOp{}.Add(gtx.Ops)
}
