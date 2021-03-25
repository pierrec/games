package ui

import (
	"image"
	"image/color"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"git.sr.ht/~pierrec/giox/widgetx"

	"github.com/pierrec/games/blocks/internal/widgets"
)

// Keymap entry names.
const (
	moveLeft = iota
	moveRight
	dropHard
	dropSoft
	rotateLeft
	rotateRight
	pauseGame
)

type keymapEntry struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}

type settings struct {
	Menu            widgets.Menu
	Padding         unit.Value
	SelectedBg      color.NRGBA
	SelectedFg      color.NRGBA
	SelectedColor   texture
	SelectedPattern texture

	keymap   []keymapEntry
	table    widgets.Table
	selected int // selected keymap entry

	listC widgetx.ClickList // list of available color textures
	listP widgetx.ClickList // list of available pattern textures
	listB layout.List       // list of blocks with the selected texture applied
}

// Menu indexes.
const (
	settingsKeymap = iota
	settingsTexture
	settingsSpace
	settingsBack
	settings_
)

func (s *settings) Key(name string) int {
	s.init()
	for i := range s.keymap {
		if s.keymap[i].Key == name {
			return i
		}
	}
	return -1
}

func (s *settings) Texture() texture {
	return s.SelectedColor | s.SelectedPattern
}

func (s *settings) saveConfig(cfg *config) {
	cfg.Keys = s.keymap
	cfg.BlockColor = s.SelectedColor
	cfg.BlockPattern = s.SelectedPattern
}

func (s *settings) loadConfig(cfg *config) {
	s.keymap = cfg.Keys
	s.SelectedColor = cfg.BlockColor
	s.SelectedPattern = cfg.BlockPattern
	// If the config file did not exist, initialize the textures.
	if s.SelectedColor.color() == transparentT {
		s.SelectedColor = redT
		s.SelectedPattern = cornerT
	}
}

func (s *settings) init() {
	if s.table.LineHeight.V == 0 {
		if s.keymap == nil {
			s.keymap = []keymapEntry{
				moveLeft:    {Text: "Move left", Key: key.NameLeftArrow},
				moveRight:   {Text: "Move right", Key: key.NameRightArrow},
				dropHard:    {Text: "Hard drop", Key: key.NameUpArrow},
				dropSoft:    {Text: "Soft drop", Key: key.NameDownArrow},
				rotateLeft:  {Text: "Rotate left", Key: "A"},
				rotateRight: {Text: "Rotate right", Key: "Z"},
				pauseGame:   {Text: "Pause", Key: key.NameEscape},
			}
		}
		s.table = widgets.Table{
			Hover:      s.Menu.Border.Color,
			LineHeight: s.Menu.Border.Width,
			LineColor:  s.Menu.Border.Color,
		}
		s.selected = -1
		s.listC = widgetx.ClickList{
			List: layout.List{Alignment: layout.Middle},
		}
		s.listP = s.listC
	}
}

func (s *settings) update(gtx layout.Context) {
	if i := s.table.Clicked(); i >= 0 {
		if s.selected == i {
			s.selected = -1
		} else {
			s.selected = i
		}
	}
	if s.selected >= 0 {
		key.InputOp{Tag: s}.Add(gtx.Ops)
		key.FocusOp{Tag: s}.Add(gtx.Ops)
	next:
		for _, ev := range gtx.Queue.Events(s) {
			k, ok := ev.(key.Event)
			if !ok || k.State != key.Release {
				continue
			}
			for i := range s.keymap {
				if s.keymap[i].Key == k.Name {
					continue next
				}
			}
			s.keymap[s.selected].Key = k.Name
		}
	}
}

func (s *settings) Layout(gtx layout.Context) layout.Dimensions {
	s.init()
	s.update(gtx)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X /= 2
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return s.Menu.Layout(gtx, settings_, func(gtx layout.Context, i int) widgets.MenuItem {
			switch i {
			case settingsKeymap:
				return widgets.MenuTitle(s.layoutKeymap, "Keyboard Map")
			case settingsTexture:
				return widgets.MenuTitle(s.layoutTextures, "Texture")
			case settingsSpace:
				return widgets.MenuSpacer(unit.Dp(20))
			case settingsBack:
				return widgets.MenuButton(func(gtx layout.Context) layout.Dimensions {
					return s.Menu.Label.Layout(gtx, "Back")
				})
			}
			return widgets.MenuItem{}
		})
	})
}

func (s *settings) layoutKeymap(gtx layout.Context) layout.Dimensions {
	return s.table.Layout(gtx, len(s.keymap), func(gtx layout.Context, idx int) layout.Dimensions {
		k := &s.keymap[idx]
		l := s.Menu.Label
		if s.selected == idx {
			l.Color = s.SelectedFg
		}
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				size := gtx.Constraints.Min
				if s.selected == idx {
					paint.FillShape(gtx.Ops, s.SelectedBg, clip.Rect{Max: size}.Op())
				}
				return layout.Dimensions{Size: size}
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:    layout.Horizontal,
					Spacing: layout.SpaceBetween,
				}.Layout(gtx,
					layout.Flexed(0.5, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Left: s.Padding,
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return l.Layout(gtx, k.Text)
							})
						})
					}),
					layout.Flexed(0.5, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Right: s.Padding,
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return l.Layout(gtx, k.Key)
							})
						})
					}),
				)
			}),
		)
	})
}

func (s *settings) layoutTextures(gtx layout.Context) layout.Dimensions {
	const selectedCell, cell = 48, 30
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			start, end := textureColors()
			if pos, _ := s.listC.Clicked(); pos >= 0 {
				s.SelectedColor = start + texture(pos)
			}
			return s.listC.Layout(gtx, int(end-start), func(gtx layout.Context, idx int, click *widget.Clickable) layout.Dimensions {
				return s.wrapTexture(gtx, func(gtx layout.Context) layout.Dimensions {
					t := start + texture(idx)
					xy := cell
					if t == s.SelectedColor || s.listC.Hovered(idx) {
						xy = selectedCell
					}
					gtx.Constraints = layout.Exact(image.Point{X: xy, Y: xy})
					return t.Layout(gtx, white)
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if pos, _ := s.listP.Clicked(); pos >= 0 {
				s.SelectedPattern = texturePattern(pos)
			}
			return s.listP.Layout(gtx, texturePatterns(), func(gtx layout.Context, idx int, click *widget.Clickable) layout.Dimensions {
				return s.wrapTexture(gtx, func(gtx layout.Context) layout.Dimensions {
					t := texturePattern(idx)
					xy := cell
					if t == s.SelectedPattern || s.listP.Hovered(idx) {
						xy = selectedCell
					}
					if t.gradient() == uniformT {
						t |= blackT
					} else {
						t |= s.SelectedColor
					}
					gtx.Constraints = layout.Exact(image.Point{X: xy, Y: xy})
					return t.Layout(gtx, white)
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		}),
	)
}

func (s *settings) wrapTexture(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return widget.Border{
			Color: s.Menu.Border.Color,
			Width: s.Menu.Border.Width.Scale(0.5),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(s.Menu.Border.Width).Layout(gtx, w)
		})
	})
}
