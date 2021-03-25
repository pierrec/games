package widgets

import (
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
)

type Button struct {
	Hover  color.NRGBA
	Border widget.Border
}

func (b Button) Layout(gtx layout.Context, click *widget.Clickable, w layout.Widget) layout.Dimensions {
	border := b.Border
	if click.Hovered() {
		border.Color = b.Hover
	}
	return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		csX := gtx.Constraints.Min.X
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				defer op.Save(gtx.Ops).Load()
				size := gtx.Constraints.Min
				r := f32.Rectangle{Max: layout.FPt(size)}
				rr := float32(gtx.Metric.Px(b.Border.CornerRadius))
				clip.UniformRRect(r, rr).Add(gtx.Ops)
				paint.LinearGradientOp{
					Stop2:  f32.Point{Y: float32(size.Y)},
					Color2: b.Border.Color,
				}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				return layout.Dimensions{Size: size}
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = csX
				return layout.Center.Layout(gtx, w)
			}),
			layout.Expanded(click.Layout),
		)
	})
}
