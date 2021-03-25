package widgets

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
)

type Label struct {
	Color  color.NRGBA
	Shaper text.Shaper
	Font   text.Font
	Size   unit.Value
	Inset  layout.Inset
	label  widget.Label
}

func (l Label) Layout(gtx layout.Context, txt string) layout.Dimensions {
	return l.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer op.Save(gtx.Ops).Load()
		paint.ColorOp{Color: l.Color}.Add(gtx.Ops)
		return l.label.Layout(gtx, l.Shaper, l.Font, l.Size, txt)
	})
}
