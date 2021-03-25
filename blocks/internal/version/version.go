package version

import (
	_ "embed"
)

//go:generate go run gen.go

//go:embed short.txt
var Short string

//go:embed long.txt
var Long string
