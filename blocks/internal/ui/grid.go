package ui

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
)

type grid struct {
	Background color.NRGBA
	cellSize   image.Point
	data       [][]texture
}

// Init initializes the grid with size as the available space.
func (g *grid) Init(width, height int) {
	g.data = nil
	g.Resize(width, height)
}

func (g *grid) Resize(width, height int) {
	if len(g.data) == 0 {
		g.data = make([][]texture, height)
		textures := make([]texture, width*height)
		for i := range g.data {
			g.data[i] = textures[:width]
			textures = textures[width:]
		}
		return
	}
	textures := g.data[0][:cap(g.data[0])]
	wh := width * height
	sz := g.Size()
	w := sz.X
	if width < w {
		w = width
	}
	old := textures
	if len(textures) < wh {
		textures = make([]texture, wh)
	}
	g.data = append(g.data[:0], make([][]texture, height)...)
	for i := range g.data {
		g.data[i] = textures[:width]
		copy(textures, old[:w])
		textures = textures[width:]
	}
}

// Size returns the width and height of the grid as a point.
func (g *grid) Size() image.Point {
	if len(g.data) == 0 {
		return image.Point{}
	}
	return image.Pt(len(g.data[0]), len(g.data))
}

// CellSize returns the width and height of a grid's cell as a point.
func (g *grid) CellSize() image.Point {
	return g.cellSize
}

// SetCellSize updates the grid cell's size.
func (g *grid) SetCellSize(size image.Point) {
	g.cellSize = size
}

func (g *grid) Clear() {
	textures := g.data[0][:cap(g.data[0])]
	copy(textures, make([]texture, len(textures)))
}

// Slice returns a subset of g that shares its cells,
// so that any cell change to the returned grid changes g's.
func (g *grid) Slice(min, max image.Point) *grid {
	if len(g.data) == 0 {
		return g
	}
	gg := *g
	var data [][]texture
	if min.X == 0 && max.X == len(g.data[0]) {
		// Slice lines only, avoid an alloc.
		data = g.data[min.Y:max.Y]
	} else {
		data = make([][]texture, max.Y-min.Y)
		for y := min.Y; y < max.Y; y++ {
			data[y] = g.data[y][min.X:max.X]
		}
	}
	gg.data = data
	return &gg
}

func (g *grid) Set(x, y int, t texture) {
	g.data[y][x] = t
}

func (g *grid) Get(x, y int) texture {
	return g.data[y][x]
}

func (g *grid) SetLine(x, y int, t ...texture) {
	copy(g.data[y][x:], t)
}

func (g *grid) Layout(gtx layout.Context) layout.Dimensions {
	if len(g.data) == 0 {
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
	defer op.Save(gtx.Ops).Load()
	// Display textures.
	gtxT := gtx
	gtxT.Constraints = layout.Exact(g.cellSize)
	var size image.Point
	for _, row := range g.data {
		var x, y int
		for _, t := range row {
			dims := t.Layout(gtxT, g.Background)
			x += dims.Size.X
			y = max(y, dims.Size.Y)
			op.Offset(f32.Point{
				X: float32(dims.Size.X),
			}).Add(gtx.Ops)
		}
		size.X = max(size.X, x)
		size.Y += y
		op.Offset(f32.Point{
			X: -float32(x),
			Y: float32(y),
		}).Add(gtx.Ops)
	}

	size.Y = max(size.Y, gtx.Constraints.Min.Y)
	return layout.Dimensions{Size: size}
}

func (g *grid) Fill(t texture) {
	for _, row := range g.data {
		for x := range row {
			row[x] = t
		}
	}
}

func (g *grid) String() string {
	buf := new(strings.Builder)

	for _, row := range g.data {
		for _, t := range row {
			buf.WriteString(t.String())
		}
		buf.WriteString("\n")
	}

	return buf.String()
}
