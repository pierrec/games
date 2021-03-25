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

type menuItemKind uint8

const (
	menuNoTitle menuItemKind = iota
	menuButton
	menuTitle
	menuSpacer
)

type MenuItem struct {
	kind   menuItemKind
	widget layout.Widget
	title  string
	v      unit.Value
}

type menuClick struct {
	idx   int
	click widget.Clickable
}

// Menu lays out a vertical set of widgets and adds a border around each of them.
// Borders will all have their width set to the widest widget.
type Menu struct {
	List   layout.List
	Label  Label
	Hover  color.NRGBA
	Border widget.Border
	clicks []menuClick
}

type MenuElement func(layout.Context, int) MenuItem

func MenuNoTitle(w layout.Widget) MenuItem {
	return MenuItem{
		kind:   menuNoTitle,
		widget: w,
	}
}

func MenuTitle(w layout.Widget, title string) MenuItem {
	return MenuItem{
		kind:   menuTitle,
		widget: w,
		title:  title,
	}
}

func MenuSpacer(v unit.Value) MenuItem {
	return MenuItem{
		kind: menuSpacer,
		v:    v,
	}
}

func MenuButton(w layout.Widget) MenuItem {
	return MenuItem{
		kind:   menuButton,
		widget: w,
	}
}

func (it *MenuItem) isButton() bool {
	return it.kind == menuButton
}

func (m *Menu) Clicked() (idx int) {
	for i := range m.clicks {
		if b := &m.clicks[i]; b.click.Clicked() {
			return b.idx
		}
	}
	return -1
}

func (m *Menu) Hovered(idx int) bool {
	if b := m.button(idx); b != nil {
		return b.Hovered()
	}
	return false
}

func (m *Menu) button(idx int) *widget.Clickable {
	for i := range m.clicks {
		if c := &m.clicks[i]; c.idx == idx {
			return &c.click
		}
	}
	return nil
}

func (m *Menu) init(n int) {
	if cn := len(m.clicks); cn < n {
		m.clicks = append(m.clicks, make([]menuClick, n-cn)...)
	}
}

// Layout lays out the widgets with a button if set as such, in the center of the current container.
func (m *Menu) Layout(gtx layout.Context, n int, el MenuElement) layout.Dimensions {
	m.init(n)
	var ci int
	return m.List.Layout(gtx, n, func(gtx layout.Context, idx int) layout.Dimensions {
		defer op.Save(gtx.Ops).Load()
		item := el(gtx, idx)
		switch item.kind {
		case menuButton:
			mclick := &m.clicks[ci]
			mclick.idx = idx
			click := &mclick.click
			ci++
			return Button{
				Hover:  m.Hover,
				Border: m.Border,
			}.Layout(gtx, click, item.widget)
		case menuSpacer:
			v := gtx.Metric.Px(item.v)
			return layout.Dimensions{Size: image.Pt(v, v)}
		}
		defer op.Save(gtx.Ops).Load()
		rec := op.Record(gtx.Ops)
		var dims layout.Dimensions
		switch item.kind {
		case menuNoTitle:
			dims = layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return item.widget(gtx)
			})
		case menuTitle:
			dims = layout.Flex{
				Axis: m.List.Axis,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					csX := gtx.Constraints.Min.X
					return layout.Stack{}.Layout(gtx,
						layout.Expanded(func(gtx layout.Context) layout.Dimensions {
							defer op.Save(gtx.Ops).Load()
							size := gtx.Constraints.Min
							clip.Rect{Max: size}.Add(gtx.Ops)
							paint.LinearGradientOp{
								Stop2:  f32.Point{Y: float32(size.Y)},
								Color2: m.Border.Color,
							}.Add(gtx.Ops)
							paint.PaintOp{}.Add(gtx.Ops)
							return layout.Dimensions{Size: size}
						}),
						layout.Stacked(func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min.X = csX
							return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return m.Label.Layout(gtx, item.title)
							})
						}),
					)
				}),
				layout.Rigid(item.widget),
			)
		}
		c := rec.Stop()
		return m.Border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			r := f32.Rectangle{Max: layout.FPt(dims.Size)}
			rr := float32(gtx.Metric.Px(m.Border.CornerRadius))
			clip.UniformRRect(r, rr).Add(gtx.Ops)
			c.Add(gtx.Ops)
			return dims
		})
	})
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
