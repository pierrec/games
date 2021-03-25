// Code generated by "stringer -type texture -linecomment -output texture_string.go"; DO NOT EDIT.

package ui

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[transparentT-0]
	_ = x[invisibleT-1]
	_ = x[whiteT-2]
	_ = x[blackT-3]
	_ = x[redT-4]
	_ = x[orangeT-5]
	_ = x[yellowT-6]
	_ = x[greenT-7]
	_ = x[blueT-8]
	_ = x[indigoT-9]
	_ = x[violetT-10]
	_ = x[_colorT-11]
	_ = x[giologo-12]
	_ = x[_imgT-13]
	_ = x[uniformT-1024]
	_ = x[squareT-2048]
	_ = x[hollowT-3072]
	_ = x[cornerT-4096]
	_ = x[pyramidT-5120]
	_ = x[gradientNT-6144]
	_ = x[gradientET-7168]
	_ = x[gradientST-8192]
	_ = x[gradientWT-9216]
	_ = x[gradientNWT-10240]
	_ = x[gradientNET-11264]
	_ = x[gradientSET-12288]
	_ = x[gradientSWT-13312]
	_ = x[_patternT-14336]
}

const _texture_name = "T_WbROYGBIVcolgiologoimguniformTsquareThollowTcornerTpyramidTgradientNTgradientETgradientSTgradientWTgradientNWTgradientNETgradientSETgradientSWT_patternT"

var _texture_map = map[texture]string{
	0:     _texture_name[0:1],
	1:     _texture_name[1:2],
	2:     _texture_name[2:3],
	3:     _texture_name[3:4],
	4:     _texture_name[4:5],
	5:     _texture_name[5:6],
	6:     _texture_name[6:7],
	7:     _texture_name[7:8],
	8:     _texture_name[8:9],
	9:     _texture_name[9:10],
	10:    _texture_name[10:11],
	11:    _texture_name[11:14],
	12:    _texture_name[14:21],
	13:    _texture_name[21:24],
	1024:  _texture_name[24:32],
	2048:  _texture_name[32:39],
	3072:  _texture_name[39:46],
	4096:  _texture_name[46:53],
	5120:  _texture_name[53:61],
	6144:  _texture_name[61:71],
	7168:  _texture_name[71:81],
	8192:  _texture_name[81:91],
	9216:  _texture_name[91:101],
	10240: _texture_name[101:112],
	11264: _texture_name[112:123],
	12288: _texture_name[123:134],
	13312: _texture_name[134:145],
	14336: _texture_name[145:154],
}

func (i texture) String() string {
	if str, ok := _texture_map[i]; ok {
		return str
	}
	return "texture(" + strconv.FormatInt(int64(i), 10) + ")"
}
