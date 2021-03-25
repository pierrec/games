package ui

import (
	"image/color"
	"strconv"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"

	"github.com/pierrec/games/blocks/internal/widgets"
)

type score struct {
	Label        widgets.Label
	Padding      unit.Value
	Level        int
	LineColor    color.NRGBA
	LineHeight   unit.Value
	LineOverflow unit.Value
	AnimBg       color.NRGBA

	data  [score_]scoreData
	table widgets.Table

	clears int
}

type scoreData struct {
	animate bool
	anim    widgets.Anim
	text    string
	val     int
	height  int
}

const (
	scoreTotal = iota
	scoreLines
	scoreLevel
	scoreLine1
	scoreLine2
	scoreLine3
	scoreLine4
	score_
)

var scoreFields = [...]scoreData{
	scoreTotal: {text: "SCORE"},
	scoreLines: {text: "LINES"},
	scoreLevel: {text: "LEVEL"},
	scoreLine1: {text: "1 LINE"},
	scoreLine2: {text: "2 LINES"},
	scoreLine3: {text: "3 LINES"},
	scoreLine4: {text: "4 LINES"},
}

// https://tetris.wiki/Tetris_(NES,_Nintendo)
func gameGravity(l int) time.Duration {
	g := 1
	switch {
	case l <= 8:
		g = 48 - 5*l
	case l == 9:
		g = 6
	case l <= 12:
		g = 5
	case l <= 15:
		g = 4
	case l <= 18:
		g = 3
	case l <= 28:
		g = 2
	}
	// 30 frames ~ 1500ms
	return time.Duration(g*1500/30) * time.Millisecond
}

// https://tetris.wiki/Scoring#Original_Nintendo_scoring_system
func (s *score) NewBlock(softDrop int) {
	s.data[scoreTotal].val += softDrop
}

func (s *score) NewLines(num int) (newLevel bool) {
	points := [4]int{40, 100, 300, 1200}
	total := s.data[scoreTotal].val
	pts := points[num-1] * (s.data[scoreLevel].val + 1)
	s.data[scoreTotal].val += pts
	// Flash the total score every 10k points.
	if threshold := 10000; (total+pts)/threshold > total/threshold {
		s.data[scoreTotal].animate = true
	}
	s.data[scoreLines].val += num
	s.data[scoreLine1+num-1].val++
	// Level change check.
	clears := s.clears + num
	startLevel := s.data[scoreLevel].val
	if clears >= (startLevel*10)+10 || clears >= max(100, (startLevel*10)-50) {
		// Change level: make it flash.
		s.data[scoreLevel].val++
		s.clears = 0
		s.data[scoreLevel].animate = true
		return true
	}
	s.clears = clears
	return false
}

func (s *score) CurrentLevel() int {
	s.init()
	return s.data[scoreLevel].val
}

func (s *score) Scores() []scoreData {
	return s.data[:]
}

func (s *score) label(gtx layout.Context, x, y int) widgets.Label {
	d := &s.data[y]
	if d.animate {
		// Flash the line for 10*200ms.
		d.animate = false
		d.anim.Start(0, 10)
	}
	if !d.anim.Animating() {
		return s.Label
	}
	v := d.anim.Value()
	if x == 0 {
		// Only move the animation forward on the first table column
		// so that subsequent ones are also animated.
		v = d.anim.Animate(gtx)
	}
	if v%2 == 0 {
		return s.Label
	}
	l := s.Label
	l.Color = s.AnimBg
	return l
}

func (s *score) anim(pos int) (int, time.Duration) {
	return pos + 1, 200 * time.Millisecond
}

func (s *score) init() {
	if s.table.LineHeight.V == 0 {
		s.data = scoreFields
		s.data[scoreLevel].val = s.Level
		for i := range s.data {
			s.data[i].anim.Next = s.anim
		}
		s.table = widgets.Table{
			LineColor:  s.LineColor,
			LineHeight: s.LineHeight,
		}
	}
}

func (s *score) Layout(gtx layout.Context) layout.Dimensions {
	s.init()
	defer func(h unit.Value) { s.table.LineHeight = h }(s.table.LineHeight)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return s.table.Layout(gtx, len(s.data), func(gtx layout.Context, idx int) layout.Dimensions {
			line := s.data[idx]
			pad := s.LineOverflow
			if idx == len(s.data)-1 {
				// Dont draw the last line.
				s.table.LineHeight = unit.Value{}
			}
			return layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceBetween,
			}.Layout(gtx,
				layout.Flexed(0.5, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: pad}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							l := s.label(gtx, 0, idx)
							return l.Layout(gtx, line.text)
						})
					})
				}),
				layout.Flexed(0.5, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Right: pad}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							l := s.label(gtx, 1, idx)
							return l.Layout(gtx, strconv.Itoa(line.val))
						})
					})
				}),
			)
		})
	})
}
