package ui

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io/fs"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type texture -linecomment -output texture_string.go

//go:embed textures
var images embed.FS

var imgTextures []paint.ImageOp

func init() {
	root := "textures"
	dirs, err := fs.ReadDir(images, root)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	for _, f := range dirs {
		if f.IsDir() {
			continue
		}
		name := fmt.Sprintf("%s/%s", root, f.Name())
		bts, err := fs.ReadFile(images, name)
		if err != nil {
			panic(err)
		}
		buf.Reset()
		buf.Write(bts)
		img, _, err := image.Decode(&buf)
		if err != nil {
			panic(err)
		}
		imgTextures = append(imgTextures, paint.NewImageOp(img))
	}
}

// texture defines a cell content:
// - uniform color
// - image (scaled)
// - pattern (values above 1<<10-1) and a color
//
// This gives 2^10 possible colors and images, and 2^6 patterns.
type texture uint16

func textureColors() (start, end texture) {
	return invisibleT + 1, _colorT - 1
}

// texturePatterns returns the number of patterns excluding all
// but the first gradient.
func texturePatterns() int {
	return int(gradientNT-uniformT)>>texturePatternBits + 1
}

// texturePattern returns the nth pattern.
func texturePattern(n int) texture {
	return uniformT + texture(n<<texturePatternBits)
}

const (
	transparentT texture = iota // T
	invisibleT                  // _
	whiteT                      // W
	blackT                      // b
	redT                        // R
	orangeT                     // O
	yellowT                     // Y
	greenT                      // G
	blueT                       // B
	indigoT                     // I
	violetT                     // V
	_colorT                     // col
	// Image textures are loaded in file name order and
	// attributed the next free texture value.
	giologo
	_imgT // img
	blurT = 1 << 15
)

const (
	texturePatternBits = 10
	textureColorMask   = 1<<texturePatternBits - 1
)

// Texture patterns.
const (
	uniformT texture = (iota + 1) << texturePatternBits
	squareT
	hollowT
	cornerT
	pyramidT
	gradientNT
	gradientET
	gradientST
	gradientWT
	gradientNWT
	gradientNET
	gradientSET
	gradientSWT
	_patternT
)

// color extracts the color from the texture.
func (t texture) color() texture {
	return t & textureColorMask
}

// pattern extracts the pattern from the texture.
func (t texture) pattern() texture {
	if p := t.unblur() &^ textureColorMask; p >= uniformT {
		return p.unblur()
	}
	return uniformT
}

// gradient extracts the gradient from the texture.
func (t texture) gradient() texture {
	if p := t.pattern(); p >= gradientNT {
		return p
	}
	return uniformT
}

// blur returns a color which alpha channel is set to 128.
func (t texture) blur() texture {
	return t | blurT
}

// unblur returns a color which alpha channel is set to 255.
func (t texture) unblur() texture {
	return t &^ blurT
}

func (t texture) nrgba() (c color.NRGBA) {
	switch t.color() {
	case whiteT:
		c = white
	case blackT:
		c = black
	case redT:
		c = red
	case orangeT:
		c = orange
	case yellowT:
		c = yellow
	case greenT:
		c = green
	case blueT:
		c = blue
	case indigoT:
		c = indigo
	case violetT:
		c = violet
	}
	if t&blurT > 0 {
		c.A = 128
	}
	return
}

func (t texture) layoutColor(gtx layout.Context) layout.Dimensions {
	switch t.color() {
	case transparentT:
		return layout.Dimensions{Size: gtx.Constraints.Min}
	case invisibleT:
		return layout.Dimensions{}
	}
	col := t.nrgba()
	r := clip.Rect{Max: gtx.Constraints.Min}.Op()
	paint.FillShape(gtx.Ops, col, r)
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func (t texture) layoutImage(gtx layout.Context) layout.Dimensions {
	defer op.Save(gtx.Ops).Load()
	size := gtx.Constraints.Min
	iOp := imgTextures[t-(_colorT+1)]
	iOp.Add(gtx.Ops)
	sz := layout.FPt(iOp.Size())
	origin := f32.Point{}
	factor := f32.Point{
		X: float32(size.X) / sz.X,
		Y: float32(size.Y) / sz.Y,
	}
	op.Affine(f32.Affine2D{}.Scale(origin, factor)).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size}
}

func (t texture) layoutPattern(gtx layout.Context, bg color.NRGBA) layout.Dimensions {
	defer op.Save(gtx.Ops).Load()
	size := gtx.Constraints.Min
	size32 := layout.FPt(size)
	pt := layout.FPt(size.Div(8))

	pattern := t.pattern()
	col := t.nrgba()
	clip.Rect{Max: size}.Add(gtx.Ops)

	height := pt.X
	orig := f32.Pt(height, height)

	switch pattern {
	case uniformT:
		clip.Rect{Max: size}.Add(gtx.Ops)
		paint.Fill(gtx.Ops, col)
	case gradientNT:
		paint.LinearGradientOp{
			Stop1:  f32.Point{Y: float32(size.Y)},
			Color2: col,
			Stop2:  f32.Point{Y: float32(size.Y) / 2},
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	case gradientST:
		paint.LinearGradientOp{
			Stop2:  f32.Point{Y: float32(size.Y) / 2},
			Color2: col,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	case gradientWT:
		paint.LinearGradientOp{
			Stop1:  f32.Point{X: float32(size.X)},
			Stop2:  f32.Point{X: float32(size.X) / 2},
			Color2: col,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	case gradientET:
		paint.LinearGradientOp{
			Stop1:  f32.Point{X: float32(size.X) / 2},
			Color1: col,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	case gradientNWT:
		paint.LinearGradientOp{
			Stop1:  f32.Point{Y: float32(size.Y)},
			Color2: col,
			Stop2:  f32.Point{Y: float32(size.Y) / 2},
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		paint.LinearGradientOp{
			Stop1:  f32.Point{X: float32(size.X)},
			Stop2:  f32.Point{X: float32(size.X) / 2},
			Color2: col,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	case gradientNET:
		paint.LinearGradientOp{
			Stop1:  f32.Point{Y: float32(size.Y)},
			Color2: col,
			Stop2:  f32.Point{Y: float32(size.Y) / 2},
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		paint.LinearGradientOp{
			Stop1:  f32.Point{X: float32(size.X) / 2},
			Color1: col,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	case gradientSWT:
		paint.LinearGradientOp{
			Stop2:  f32.Point{Y: float32(size.Y) / 2},
			Color2: col,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		paint.LinearGradientOp{
			Stop1:  f32.Point{X: float32(size.X)},
			Stop2:  f32.Point{X: float32(size.X) / 2},
			Color2: col,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	case gradientSET:
		paint.LinearGradientOp{
			Stop2:  f32.Point{Y: float32(size.Y) / 2},
			Color2: col,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		paint.LinearGradientOp{
			Stop1:  f32.Point{X: float32(size.X) / 2},
			Color1: col,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
	case squareT:
		paint.Fill(gtx.Ops, col)
		clip.RRect{
			Rect: f32.Rectangle{
				Min: orig,
				Max: size32.Sub(orig),
			},
		}.Add(gtx.Ops)
		paint.Fill(gtx.Ops, bg)

		orig = orig.Mul(2)
		clip.RRect{
			Rect: f32.Rectangle{
				Min: orig,
				Max: size32.Sub(orig),
			},
		}.Add(gtx.Ops)
		paint.Fill(gtx.Ops, col)
	case hollowT:
		paint.Fill(gtx.Ops, col)
		orig = orig.Mul(2)
		clip.RRect{
			Rect: f32.Rectangle{
				Min: orig,
				Max: size32.Sub(orig),
			},
		}.Add(gtx.Ops)
		paint.Fill(gtx.Ops, bg)
	case cornerT:
		paint.Fill(gtx.Ops, col)
		clip.RRect{
			Rect: f32.Rectangle{
				Min: orig,
				Max: size32.Sub(orig),
			},
		}.Add(gtx.Ops)
		paint.Fill(gtx.Ops, bg)

		orig = orig.Mul(2)
		clip.RRect{
			Rect: f32.Rectangle{
				Min: orig,
				Max: size32,
			},
		}.Add(gtx.Ops)
		paint.Fill(gtx.Ops, col)
	case pyramidT:
		paint.Fill(gtx.Ops, col)
		step := f32.Pt(height/2, height/2)
		for f := true; orig.X < size32.X/2; f = !f {
			clip.RRect{
				Rect: f32.Rectangle{
					Min: orig,
					Max: size32.Sub(orig),
				},
			}.Add(gtx.Ops)
			if f {
				paint.Fill(gtx.Ops, bg)
			} else {
				paint.Fill(gtx.Ops, col)
			}
			orig = orig.Add(step)
		}
	default:
		msg := fmt.Sprintf("unknown texture: %s", t)
		panic(msg)
	}
	return layout.Dimensions{Size: size}
}

// The minimum constraints must be set to the cell size.
func (t texture) Layout(gtx layout.Context, bg color.NRGBA) layout.Dimensions {
	if t.pattern() == uniformT {
		switch c := t.color(); {
		case c < _colorT:
			return t.layoutColor(gtx)
		case c < _imgT:
			return c.layoutImage(gtx)
		}
	}
	return t.layoutPattern(gtx, bg)
}
