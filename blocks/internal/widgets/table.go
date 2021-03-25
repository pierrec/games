package widgets

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

type Table struct {
	Hover      color.NRGBA
	LineColor  color.NRGBA
	LineHeight unit.Value
	clicks     []widget.Clickable
}

type TableRow func(layout.Context, int) layout.Dimensions

type TableColumn struct {
	weight float32
	widget layout.Widget
}

func (t *Table) Clicked() int {
	for i := range t.clicks {
		if t.clicks[i].Clicked() {
			return i
		}
	}
	return -1
}

func (t *Table) init(n int) {
	if cn := len(t.clicks); cn < n {
		t.clicks = append(t.clicks, make([]widget.Clickable, n-cn)...)
	}
}

func (t *Table) Layout(gtx layout.Context, rows int, el TableRow) layout.Dimensions {
	defer op.Save(gtx.Ops).Load()
	t.init(rows)
	var size image.Point
	for y := 0; y < rows; y++ {
		click := &t.clicks[y]
		dims := layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				size := gtx.Constraints.Min
				if click.Hovered() {
					paint.FillShape(gtx.Ops, t.Hover, clip.Rect{Max: size}.Op())
				}
				return layout.Dimensions{Size: size}
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				rec := op.Record(gtx.Ops)
				gtx.Constraints.Min.Y = 0
				dims := el(gtx, y)
				call := rec.Stop()
				if t.LineHeight.V == 0 {
					call.Add(gtx.Ops)
					return dims
				}
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						call.Add(gtx.Ops)
						return dims
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						height := gtx.Metric.Px(t.LineHeight)
						size := image.Point{X: dims.Size.X, Y: height}
						paint.FillShape(gtx.Ops, t.LineColor, clip.Rect{Max: size}.Op())
						return layout.Dimensions{Size: size}
					}),
				)
			}),
			layout.Expanded(click.Layout),
		)
		size.X = max(size.X, dims.Size.X)
		size.Y += dims.Size.Y
		op.Offset(f32.Point{Y: float32(dims.Size.Y)}).Add(gtx.Ops)
	}
	return layout.Dimensions{Size: size}
}
