package ui

import (
	"sort"
	"strconv"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/pierrec/games/blocks/internal/widgets"
)

type scoreEntry struct {
	Player string      `json:"player"`
	Score  [score_]int `json:"score"`
}

type scoreboard struct {
	Menu     widgets.Menu
	Padding  unit.Value
	state    uint8
	table    widgets.Table
	data     [10 + 1]scoreEntry // keep the last 10 entries and 1 as scratch
	ed       widget.Editor
	newScore int
}

const (
	scoreboardShow = iota
	scoreboardPlayer
)

// Menu indexes.
const (
	scoreboardData = iota
	scoreboardSpace
	scoreboardBack
	scoreboard_
)

func (s *scoreboard) totalAt(i int) int {
	return s.data[i].Score[scoreTotal]
}

// NewScore returns whether or not the score makes it in the board.
func (s *scoreboard) NewScore(score []scoreData) bool {
	total := score[scoreTotal].val
	if total > 0 {
		for i := range s.data {
			switch t := s.totalAt(i); {
			case t < total:
			case t == total && i < len(s.data)-1:
			default:
				continue
			}
			// Add the new score.
			s.state = scoreboardPlayer
			var e scoreEntry
			for i, d := range score {
				e.Score[i] = d.val
			}
			s.data[len(s.data)-1] = e
			sort.SliceStable(s.data[:], func(i, j int) bool {
				return s.totalAt(j) < s.totalAt(i) // reverse
			})
			s.newScore = i
			return true
		}
	}
	return false
}

func (s *scoreboard) saveConfig(cfg *config) {
	cfg.Scores = s.data[:len(s.data)-1]
}

func (s *scoreboard) loadConfig(cfg *config) {
	copy(s.data[:], cfg.Scores)
}

func (s *scoreboard) init() {
	if s.ed.SingleLine == false {
		s.ed = widget.Editor{
			Alignment:  text.Middle,
			Submit:     true,
			SingleLine: true,
		}
		s.table = widgets.Table{
			LineHeight: s.Menu.Border.Width,
			LineColor:  s.Menu.Border.Color,
		}
	}
}

func (s *scoreboard) update() {
	switch s.state {
	case scoreboardShow:
	case scoreboardPlayer:
		// Limit the player name to 10 characters.
		if s.ed.Len() > 10 {
			s.ed.Delete(-1)
		}
		for _, ev := range s.ed.Events() {
			e, ok := ev.(widget.EditorEvent)
			if !ok {
				continue
			}
			switch e := e.(type) {
			case widget.SubmitEvent:
				s.state = scoreboardShow
				s.data[s.newScore].Player = e.Text
				s.ed.SetText("")
			}
		}
	}
}

func (s *scoreboard) Layout(gtx layout.Context) layout.Dimensions {
	s.init()
	s.update()
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X /= 2
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return s.Menu.Layout(gtx, scoreboard_, func(gtx layout.Context, i int) widgets.MenuItem {
			switch i {
			case scoreboardData:
				return widgets.MenuTitle(s.layoutScores, "Best Scores")
			case scoreboardSpace:
				return widgets.MenuSpacer(unit.Dp(20))
			case scoreboardBack:
				return widgets.MenuButton(func(gtx layout.Context) layout.Dimensions {
					return s.Menu.Label.Layout(gtx, "Back")
				})
			}
			return widgets.MenuItem{}
		})
	})
}

func (s *scoreboard) layoutScores(gtx layout.Context) layout.Dimensions {
	noScore := true
	for _, sc := range s.data {
		if sc != (scoreEntry{}) {
			noScore = false
			break
		}
	}
	if noScore {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return s.Menu.Label.Layout(gtx, "No score")
		})
	}
	data := s.data[:len(s.data)-1] // only display the first 10 entries
	n := len(data)
	return s.table.Layout(gtx, n, func(gtx layout.Context, idx int) layout.Dimensions {
		line := data[idx]
		isPlayer := s.state == scoreboardPlayer && s.newScore == idx
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				size := gtx.Constraints.Min
				if isPlayer {
					paint.FillShape(gtx.Ops, s.Menu.Label.Color, clip.Rect{Max: size}.Op())
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
								if !isPlayer {
									return s.Menu.Label.Layout(gtx, line.Player)
								}
								l := s.Menu.Label
								paint.ColorOp{Color: s.Menu.Border.Color}.Add(gtx.Ops)
								s.ed.PaintText(gtx)
								dims := s.ed.Layout(gtx, l.Shaper, l.Font, l.Size)
								s.ed.PaintCaret(gtx)
								s.ed.Focus()
								return dims
							})
						})
					}),
					layout.Flexed(0.5, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Right: s.Padding,
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								l := s.Menu.Label
								if isPlayer {
									l.Color = s.Menu.Border.Color
								}
								return l.Layout(gtx, strconv.Itoa(s.totalAt(idx)))
							})
						})
					}),
				)
			}),
		)
	})
}
