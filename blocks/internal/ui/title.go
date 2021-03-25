package ui

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"

	"github.com/pierrec/games/blocks/internal/widgets"
)

const TitleName = "Blocks"

var titleData = map[rune][][]texture{
	'B': {
		{redT, redT, redT, transparentT, transparentT},
		{redT, transparentT, redT, transparentT, transparentT},
		{redT, transparentT, redT, transparentT, transparentT},
		{redT, redT, redT, redT, redT},
		{redT, transparentT, transparentT, redT, redT},
		{redT, transparentT, transparentT, redT, redT},
		{redT, redT, redT, redT, redT},
	},
	'l': {
		{redT},
		{redT},
		{redT},
		{redT},
		{redT},
		{redT},
		{redT},
	},
	'o': {
		{transparentT, transparentT, transparentT, transparentT},
		{transparentT, transparentT, transparentT, transparentT},
		{transparentT, transparentT, transparentT, transparentT},
		{redT, redT, redT, redT},
		{redT, transparentT, redT, redT},
		{redT, transparentT, redT, redT},
		{redT, redT, redT, redT},
	},
	'c': {
		{transparentT, transparentT, transparentT},
		{transparentT, transparentT, transparentT},
		{transparentT, transparentT, transparentT},
		{redT, redT, redT},
		{redT, transparentT, transparentT},
		{redT, transparentT, transparentT},
		{redT, redT, redT},
	},
	'k': {
		{redT, transparentT, transparentT},
		{redT, transparentT, redT},
		{redT, transparentT, redT},
		{redT, redT, transparentT},
		{redT, transparentT, redT},
		{redT, transparentT, redT},
		{redT, transparentT, redT},
	},
	's': {
		{transparentT, transparentT, transparentT},
		{transparentT, transparentT, transparentT},
		{redT, redT, redT},
		{redT, transparentT, transparentT},
		{redT, redT, redT},
		{transparentT, transparentT, redT},
		{redT, redT, redT},
	},
}

type title struct {
	Gravity    time.Duration
	Background color.NRGBA
	Texture    texture
	state      uint8
	grid       grid
	anim       widgets.Anim
}

const (
	titleNone = iota
	titleWait
	titleFlash
	titleDown
)

func (t *title) init() {
	if t.grid.Size() == (image.Point{}) {
		var width, height int
		for _, c := range TitleName {
			s := titleData[c]
			width += len(s[0]) + 1
			height = len(s)
		}
		width++
		t.grid.Init(width, height)
		t.grid.Background = t.Background
		var line []texture
		for y := 0; y < height; y++ {
			line = line[:0]
			for _, c := range TitleName {
				s := titleData[c]
				line = append(line, transparentT)
				line = append(line, s[y]...)
			}
			line = append(line, transparentT)
			t.grid.SetLine(0, y, line...)
		}
	}
	// The texture has changed, update the grid.
	if t.grid.Get(0, 0) != t.Texture {
		sz := t.grid.Size()
		for y := 0; y < sz.Y; y++ {
			for x := 0; x < sz.X; x++ {
				if t.grid.Get(x, y) != transparentT {
					t.grid.Set(x, y, t.Texture)
				}
			}
		}
	}
}

func (t *title) update(gtx layout.Context) {
	// Set an upper bound for cell size so the title doesnt take most of the screen.
	cell := gtx.Constraints.Max.X / t.grid.Size().X
	cell = min(cell, 64)
	t.grid.SetCellSize(image.Pt(cell, cell))
	switch t.state {
	case titleNone:
		t.state = titleWait
		t.anim.Next = t.titleAnimWait
		t.anim.Start(0, 10)
		fallthrough
	case titleWait:
		if t.anim.Animating() {
			t.anim.Animate(gtx)
			break
		}
		t.state = titleFlash
		t.anim.Next = t.titleAnimFlash
		t.anim.Start(0, 5)
		fallthrough
	case titleFlash:
		if t.anim.Animating() {
			t.anim.Animate(gtx)
			break
		}
		t.state = titleDown
		t.anim.Next = func(i int) (int, time.Duration) {
			return i + 1, t.Gravity
		}
		t.anim.Start(0, 3)
		fallthrough
	case titleDown:
		if t.anim.Animating() {
			pos := t.anim.Animate(gtx)
			t.titleAnimDown(pos)
			break
		}
		t.state = titleNone
		op.InvalidateOp{}.Add(gtx.Ops)
	}
}

func (t *title) Layout(gtx layout.Context) layout.Dimensions {
	t.init()
	t.update(gtx)
	return t.grid.Layout(gtx)
}

func (t *title) letterPos(letter rune) (x0 int) {
	x0 = 1
	for _, c := range TitleName {
		if c == letter {
			return
		}
		x0 += len(titleData[c][0]) + 1
	}
	return -1
}

func (t *title) titleAnimWait(i int) (int, time.Duration) {
	return i + 1, 200 * time.Millisecond
}

func (t *title) titleAnimFlash(i int) (int, time.Duration) {
	xn := len(titleData['o'][0])
	yn := len(titleData['o'])
	tex := t.Texture
	if i%2 == 0 {
		tex = tex.blur()
	}
	x0 := t.letterPos('o')
	for y := 0; y < yn; y++ {
		for x := 0; x < xn; x++ {
			if t.grid.Get(x0+x, y) != transparentT {
				t.grid.Set(x0+x, y, tex)
			}
		}
	}
	return i + 1, 200 * time.Millisecond
}

func (t *title) titleAnimDown(pos int) (int, time.Duration) {
	xn := len(titleData['o'][0])
	yn := len(titleData['o'])
	x0 := t.letterPos('o')
	// Find the line where the letter starts.
	var y0 int
	for y := 0; y < yn; y++ {
		if titleData['o'][y][0] != transparentT {
			y0 = y
			break
		}
	}
	if pos == 0 {
		// Clear the previous letter.
		for y := 0; y < yn-y0; y++ {
			for x := 0; x < xn; x++ {
				t.grid.Set(x0+x, y0+y, transparentT)
			}
		}
	} else {
		// Clear the previous line.
		for x := 0; x < xn; x++ {
			t.grid.Set(x0+x, pos-1, transparentT)
		}
	}
	// Draw the letter at the new position.
	for y := 0; y < yn-y0; y++ {
		line := titleData['o'][y0+y]
		for x := 0; x < xn; x++ {
			tex := transparentT
			if line[x] != transparentT {
				tex = t.Texture
			}
			t.grid.Set(x0+x, pos+y, tex)
		}
	}
	return pos + 1, t.Gravity
}
