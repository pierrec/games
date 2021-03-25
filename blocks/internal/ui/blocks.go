// +build !cha,!flo,!luc

package ui

var blocks = []block{
	I: {
		id: I,
		data: [][]texture{
			{transparentT, transparentT, transparentT, transparentT},
			{gradientNT, gradientNT, gradientNT, gradientNT},
			{transparentT, transparentT, transparentT, transparentT},
			{transparentT, transparentT, transparentT, transparentT},
		},
		width: 4, height: 1,
	},
	J: {
		id: J,
		data: [][]texture{
			{transparentT, transparentT, transparentT},
			{gradientNT, gradientNT, gradientNET},
			{transparentT, transparentT, gradientET},
		},
		width: 3, height: 2,
	},
	L: {
		id: L,
		data: [][]texture{
			{transparentT, transparentT, transparentT},
			{gradientNWT, gradientNT, gradientNT},
			{gradientWT, transparentT, transparentT},
		},
		width: 3, height: 2,
	},
	O: {
		id: O,
		data: [][]texture{
			{gradientSET, gradientSWT},
			{gradientNET, gradientNWT},
		},
		width: 2, height: 2,
	},
	S: {
		id: S,
		data: [][]texture{
			{transparentT, transparentT, transparentT},
			{transparentT, uniformT, gradientNT},
			{gradientST, uniformT, transparentT},
		},
		width: 3, height: 2,
	},
	T: {
		id: T,
		data: [][]texture{
			{transparentT, transparentT, transparentT},
			{gradientNT, uniformT, gradientNT},
			{transparentT, gradientNT, transparentT},
		},
		width: 3, height: 2,
	},
	Z: {
		id: Z,
		data: [][]texture{
			{transparentT, transparentT, transparentT},
			{gradientNT, uniformT, transparentT},
			{transparentT, uniformT, gradientST},
		},
		width: 3, height: 2,
	},
}
