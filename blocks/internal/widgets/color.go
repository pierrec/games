package widgets

import "image/color"

func colorToNRGBA(c color.Color) color.NRGBA {
	r, g, b, a := c.RGBA()
	if a == 255 {
		return color.NRGBA{
			R: uint8(r),
			G: uint8(g),
			B: uint8(b),
			A: uint8(a),
		}
	}
	return color.NRGBA{
		R: uint8(r * a / 255),
		G: uint8(g * a / 255),
		B: uint8(b * a / 255),
		A: uint8(a),
	}
}
