package ui

import (
	"strconv"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"git.sr.ht/~pierrec/giox/layoutx"

	"github.com/pierrec/games/blocks/internal/version"
	"github.com/pierrec/games/blocks/internal/widgets"
)

type home struct {
	Menu     widgets.Menu
	Title    title
	Version  widgets.Label
	Error    error
	levels   [10]widget.Bool
	list     layoutx.ListWrap
	selected int
	errAnim  widgets.Anim
}

const (
	homeLevels = iota
	homeSpace1
	homeStartGame
	homeScoreBoard
	homeSettings
	homeSpace2
	homeQuitGame
	home_
)

func (h *home) Level() int {
	return h.selected
}

func (h *home) saveConfig(cfg *config) {
	cfg.Level = h.Level()
}

func (h *home) loadConfig(cfg *config) {
	if cfg.Level < len(h.levels) {
		h.selected = cfg.Level
	}
}

func (h *home) update() {
	if h.Error != nil && !h.errAnim.Animating() {
		h.errAnim.Next = func(i int) (int, time.Duration) {
			return i + 1, 5 * time.Second
		}
		h.errAnim.Start(0, 1)
	}
	for i := range h.levels {
		b := &h.levels[i]
		if b.Changed() && b.Value {
			h.levels[h.selected].Value = false
			h.selected = i
			return
		}
	}
}

func (h *home) Layout(gtx layout.Context) layout.Dimensions {
	h.update()
	layout.SE.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return h.Version.Layout(gtx, version.Long)
	})
	if h.errAnim.Animating() {
		switch h.errAnim.Animate(gtx) {
		case 0:
			layout.SW.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return h.Version.Layout(gtx, h.Error.Error())
			})
		case 1:
			h.Error = nil
		}
	}
	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceEvenly,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return h.Title.Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X /= 2
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return h.Menu.Layout(gtx, home_,
					func(gtx layout.Context, i int) widgets.MenuItem {
						l := h.Menu.Label
						switch i {
						case homeLevels:
							return widgets.MenuTitle(h.layoutLevels, "Select Level")
						case homeSpace1:
							return widgets.MenuSpacer(unit.Dp(40))
						case homeStartGame:
							return widgets.MenuButton(func(gtx layout.Context) layout.Dimensions {
								return l.Layout(gtx, "Start Game")
							})
						case homeScoreBoard:
							return widgets.MenuButton(func(gtx layout.Context) layout.Dimensions {
								return l.Layout(gtx, "Score Board")
							})
						case homeSettings:
							return widgets.MenuButton(func(gtx layout.Context) layout.Dimensions {
								return l.Layout(gtx, "Settings")
							})
						case homeSpace2:
							return widgets.MenuSpacer(unit.Dp(20))
						case homeQuitGame:
							return widgets.MenuButton(func(gtx layout.Context) layout.Dimensions {
								return l.Layout(gtx, "Quit Game")
							})
						}
						return widgets.MenuItem{}
					},
				)
			})
		}),
	)
}

func (h *home) layoutLevels(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return h.list.Layout(gtx, len(h.levels), func(gtx layout.Context, idx int) layout.Dimensions {
			bg := h.Menu.Border.Color
			l := h.Menu.Label
			selected := idx == h.selected
			if selected {
				l.Font.Weight = text.Bold
				l.Color, bg = bg, l.Color
			}
			lvl := &h.levels[idx]

			return layout.Stack{}.Layout(gtx,
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					size := gtx.Constraints.Min
					if selected {
						// Circled selection.
						c := size.Div(2)
						r := float32(min(c.X, c.Y)) * 1.2
						op := clip.Circle{
							Center: layout.FPt(c),
							Radius: r,
						}.Op(gtx.Ops)
						paint.FillShape(gtx.Ops, bg, op)
						size.X = max(size.X, int(r))
						size.Y = max(size.Y, int(r))
					}
					return layout.Dimensions{Size: size}
				}),
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						txt := strconv.Itoa(idx)
						return l.Layout(gtx, txt)
					})
				}),
				layout.Expanded(lvl.Layout),
			)
		}, nil)
	})
}
