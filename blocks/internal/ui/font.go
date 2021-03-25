package ui

import (
	_ "embed"

	"gioui.org/font/opentype"
	"gioui.org/text"
)

//go:embed data/PressStart2P-Regular.ttf
var fontBytes []byte

var (
	fntName        = text.Typeface("")
	fnt, _         = opentype.Parse(fontBytes)
	fontCollection = [12]text.FontFace{
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
			},
			Face: fnt,
		},
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
				Style:    text.Italic,
			},
			Face: fnt,
		},
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
				Weight:   text.Bold,
			},
			Face: fnt,
		},
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
				Variant:  "Sans",
				Style:    text.Italic,
				Weight:   text.Bold,
			},
			Face: fnt,
		},
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
				Weight:   text.Medium,
			},
			Face: fnt,
		},
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
				Style:    text.Italic,
				Weight:   text.Medium,
			},
			Face: fnt,
		},
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
				Variant:  "Mono",
			},
			Face: fnt,
		},
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
				Variant:  "Mono",
				Weight:   text.Bold,
			},
			Face: fnt,
		},
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
				Variant:  "Mono",
				Style:    text.Italic,
				Weight:   text.Bold,
			},
			Face: fnt,
		},
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
				Variant:  "Mono",
				Style:    text.Italic,
			},
			Face: fnt,
		},
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
				Variant:  "Smallcaps",
			},
			Face: fnt,
		},
		text.FontFace{
			Font: text.Font{
				Typeface: fntName,
				Variant:  "Smallcaps",
				Style:    text.Italic,
			},
			Face: fnt,
		},
	}
)
