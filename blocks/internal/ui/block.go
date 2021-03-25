package ui

import (
	"image"
	"math/rand"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
)

type blockID uint8

const (
	I blockID = iota
	J
	L
	O
	S
	T
	Z
)

type blockRotation uint8

// clockwise block rotations.
const (
	block0 blockRotation = iota
	block90
	block180
	block270
)

type block struct {
	KeyMap  func(string) int
	Texture texture
	ready   bool // whether or not the block has been laid out at least once
	id      blockID
	pos     image.Point
	rot     blockRotation
	data    [][]texture
	width   int
	height  int
}

func (a blockRotation) Next() blockRotation {
	return (a + 1) % 4
}

func (a blockRotation) Prev() blockRotation {
	return (a + 3) % 4
}

func (a blockRotation) NextGradient(t texture) texture {
	anchor := gradientNT
	if t >= gradientNWT {
		anchor = gradientNWT
	}
	g := (t.gradient() - anchor) >> texturePatternBits
	g = (g + 1) % 4
	return g<<texturePatternBits + anchor
}

// Dims returns the full dimensions of the block,
// including its padding.
func (b *block) Dims() (fullWidth, fullHeight int) {
	return len(b.data[0]), len(b.data)
}

// Init sets the block's data to the one at blocks index idx.
func (b *block) Init(idx blockID, t texture) {
	km := b.KeyMap
	*b = blocks[idx]
	b.KeyMap = km
	b.Texture = t
}

// InitRandom sets the block's data randomly.
func (b *block) InitRandom(t texture) {
	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(len(blocks))
	b.Init(blockID(idx), t)
}

func (b *block) ID() blockID {
	return b.id
}

// Width returns the block's width, without padding.
func (b *block) Width() int {
	return b.width
}

// Height returns the block's height, without padding.
func (b *block) Height() int {
	return b.height
}

// Pos returns the current block position on the grid,
// including its padding.
func (b *block) Pos() image.Point {
	return image.Pt(b.pos.X, max(0, b.pos.Y))
}

// MoveDown moves the block one line down and returns
// whether or not it was possible.
func (b *block) MoveDown(g *grid) (ok bool) {
	if !b.ready {
		// Consider that the move is successful while getting ready.
		return true
	}
	b.layout(g, true)
	b.pos.Y++
	ok = b.check(g)
	if !ok {
		b.pos.Y--
		b.layout(g, false)
	}
	return
}

func (b *block) walk(fn func(x, y int, t texture) bool) {
	xn, yn := b.Dims()
	isGradient := b.Texture.gradient() != uniformT
	for y := 0; y < yn; y++ {
		for x := 0; x < xn; x++ {
			xx, yy := x, y
			var t texture
			switch b.rot {
			case block0:
				t = b.data[y][x]
			case block90:
				t = b.data[yn-1-y][x]
				xx, yy = y, x
			case block180:
				t = b.data[yn-1-y][xn-1-x]
			case block270:
				t = b.data[y][xn-1-x]
				xx, yy = y, x
			}
			// Skip transparent textures.
			if t == transparentT {
				continue
			}
			if isGradient {
				if t.gradient() != uniformT {
					switch b.rot {
					case block0:
					case block90:
						t = b.rot.NextGradient(t)
					case block180:
						t = b.rot.NextGradient(t)
						t = b.rot.NextGradient(t)
					case block270:
						t = b.rot.NextGradient(t)
						t = b.rot.NextGradient(t)
						t = b.rot.NextGradient(t)
					}
				}
				t |= b.Texture.color()
			} else {
				t = b.Texture
			}
			if fn(xx, yy, t) {
				return
			}
		}
	}
}

// check reports whether or not all non transparent textures
// of the block do not collide with anything on the grid.
func (b *block) check(g *grid) (ok bool) {
	ok = true
	b.walk(func(x, y int, t texture) bool {
		if b.pos.Y+y < 0 || g.Get(b.pos.X+x, b.pos.Y+y) != transparentT {
			ok = false
			return true
		}
		return false
	})
	return
}

// init positions the block for the first time. It may fail.
func (b *block) init(gtx layout.Context, g *grid) (ok bool) {
	if b.ready {
		return true
	}
	// First attempt at positioning the block.
	b.ready = true
	// Make a block starts mid width.
	cols := g.Size().X
	b.pos.X = (cols - b.Width()) / 2
	b.pos.Y = 1 // the grid's first line is hidden.
	// Skip first empty lines so that the block gets displayed at the top edge.
	var skip int
	b.walk(func(x, y int, t texture) bool {
		if skip == y {
			return t != transparentT
		}
		skip = y
		return false
	})
	b.pos.Y -= skip
	return b.check(g)
}

// layout draws or clears the block on the grid.
func (b *block) layout(g *grid, clear bool) {
	b.walk(func(x, y int, t texture) bool {
		if clear {
			t = transparentT
		}
		if b.pos.Y+y >= 0 {
			g.Set(b.pos.X+x, b.pos.Y+y, t)
		}
		return false
	})
}

func (b *block) update(evs []event.Event, g *grid) (softDrops int, ok bool) {
	if len(evs) == 0 {
		b.layout(g, false)
		return 0, true
	}
	for _, ev := range evs {
		e, k := ev.(key.Event)
		// You get one event for a key press and one for its release, ignore the first one.
		if !k || e.State != key.Release {
			continue
		}
		switch b.KeyMap(e.Name) {
		case moveLeft:
			b.layout(g, true)
			b.pos.X--
			if !b.check(g) {
				b.pos.X++
			}
			b.layout(g, false)
		case moveRight:
			b.layout(g, true)
			b.pos.X++
			if !b.check(g) {
				b.pos.X--
			}
			b.layout(g, false)
		case dropHard:
			for b.MoveDown(g) {
				softDrops++
			}
			return
		case dropSoft:
			if !b.MoveDown(g) {
				return
			}
			softDrops++
		case rotateLeft:
			rot := b.rot
			b.layout(g, true)
			b.rot = b.rot.Prev()
			if !b.check(g) {
				b.rot = rot
			}
			b.layout(g, false)
		case rotateRight:
			rot := b.rot
			b.layout(g, true)
			b.rot = b.rot.Next()
			if !b.check(g) {
				b.rot = rot
			}
			b.layout(g, false)
		}
	}
	return softDrops, true
}

func (b *block) Layout(gtx layout.Context, g *grid, update func(int), over func()) layout.Dimensions {
	if !b.init(gtx, g) {
		over()
		op.InvalidateOp{}.Add(gtx.Ops)
		return layout.Dimensions{}
	}
	if softDrops, ok := b.update(gtx.Queue.Events(b), g); !ok {
		// The user could not move the block down: update the game loop.
		update(softDrops)
		op.InvalidateOp{}.Add(gtx.Ops)
		return layout.Dimensions{}
	}
	xn, yn := b.Dims()
	cell := g.CellSize()
	size := image.Point{
		X: xn * cell.X,
		Y: yn * cell.Y,
	}
	return layout.Dimensions{Size: size}
}
