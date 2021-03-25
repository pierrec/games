package ui

import (
	"fmt"
	"image"
	"strings"
	"testing"
)

func gridString(s string) string {
	return strings.Join(
		strings.Split(s, " "),
		"\n",
	) + "\n"
}

var textureMap = map[string]texture{}

func init() {
	for t := texture(0); t < _T; t++ {
		textureMap[t.String()] = t
	}
}

// Populate g with textures from data.
func gridFromString(g *grid, data string) {
	for y, row := range strings.Split(data, " ") {
		for x, s := range row {
			g.Set(x, y, textureMap[string(s)])
		}
	}
}

// Test laying then clearing blocks.
// Only the non transparent block cells must be erased.
func TestBlockLayout(t *testing.T) {
	type tcase struct {
		index blockID
		drawn string
	}
	for _, tc := range []tcase{
		{
			index: I,
			drawn: `bbbb__ ______ ______`,
		},
		{
			index: J,
			drawn: `b_____ bbb___ ______`,
		},
		{
			index: L,
			drawn: `__b___ bbb___ ______`,
		},
		{
			index: O,
			drawn: `bb____ bb____ ______`,
		},
		{
			index: S,
			drawn: `_bb___ bb____ ______`,
		},
		{
			index: T,
			drawn: `_b____ bbb___ ______`,
		},
		{
			index: Z,
			drawn: `bb____ _bb___ ______`,
		},
	} {
		t.Run(tc.index.String(), func(t *testing.T) {
			var g grid
			g.Init(6, 3)
			g.Fill(invisibleT)
			var b block
			b.Init(tc.index)
			var got, want string

			b.layout(&g, false)
			got = g.String()
			want = gridString(tc.drawn)
			if got != want {
				t.Errorf("got %q; want %q", got, want)
			}

			b.layout(&g, true)
			got = g.String()
			want = strings.ReplaceAll(gridString(tc.drawn), blackT.String(), transparentT.String())
			if got != want {
				t.Errorf("got %q; want %q", got, want)
			}
		})
	}
}

func TestBlockCollision(t *testing.T) {
	type coll struct {
		left, right, down bool
	}
	type tcase struct {
		index blockID
		pos   image.Point
		grid  []string
		res   []coll
	}
	for _, tc := range []tcase{
		{
			index: I,
			pos:   image.Pt(1, 0),
			grid: []string{
				`TbbbbT TTTTTT`,
				`_bbbb_ Ta_TTTT`,
			},
			res: []coll{
				{false, false, false},
				{true, true, true},
			},
		},
	} {
		t.Run(tc.index.String(), func(t *testing.T) {
			for gi, gs := range tc.grid {
				t.Run(fmt.Sprintf("%d", gi), func(t *testing.T) {
					var g grid
					g.Init(6, 3)
					gridFromString(&g, gs)
					var b block
					b.Init(tc.index)
					b.pos = tc.pos
					b.layout(&g, false)
					var got, want bool

					b.pos.X--
					got = b.check(&g)
					want = tc.res[gi].left
					if got != want {
						t.Errorf("left: got %v; want %v", got, want)
					}
					b.pos.X++
					b.pos.X += b.Width()
					got = b.check(&g)
					want = tc.res[gi].right
					if got != want {
						t.Errorf("right: got %v; want %v", got, want)
					}
					b.pos.X -= b.Width()

					b.pos.Y++
					got = b.check(&g)
					want = tc.res[gi].down
					if got != want {
						t.Errorf("down: got %v; want %v", got, want)
					}
					b.pos.Y--
				})
			}
		})
	}
}
