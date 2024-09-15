package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/joscherrer/dofus-manager/internal/f32color"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type SetEditor struct {
	Theme       *material.Theme
	Editor      *widget.Editor
	Save        *Clickable
	EditorInset layout.Inset
	Hint        string
}

func (s *SetEditor) NewTextBox(editor material.EditorStyle) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				rr := gtx.Dp(4)
				defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Max}, rr).Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, s.Theme.ContrastBg)
				return layout.Dimensions{Size: gtx.Constraints.Max}
			},
			func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(8).Layout(gtx, editor.Layout)
			},
		)
	}
}

func (s *SetEditor) Layout(gtx layout.Context) layout.Dimensions {
	flex := layout.Flex{Axis: layout.Horizontal}
	mEditor := material.Editor(s.Theme, s.Editor, "Profile")
	mEditor.HintColor = f32color.MulAlpha(s.Theme.Fg, 0x55)
	mEditor.SelectionColor = f32color.Hovered(s.Theme.Bg)
	textBox := s.NewTextBox(mEditor)

	saveIcon, _ := widget.NewIcon(icons.ContentSave)
	saveBtn := IconBtn(s.Theme, s.Save, saveIcon, "Save")

	saveBtnW := func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: 5}.Layout(gtx, saveBtn.Layout)
	}
	return layout.UniformInset(10).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.Y = 38
		return flex.Layout(gtx, layout.Flexed(1, textBox), layout.Rigid(saveBtnW))
	})
}
