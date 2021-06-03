package main

import (
	"fmt"
	"os"

	"gioui.org/app"
	"gioui.org/unit"

	"github.com/pierrec/games/blocks/internal/ui"
)

func main() {
	go func() {
		w := app.NewWindow(
			app.Size(unit.Dp(500), unit.Dp(600)),
			app.Title(ui.TitleName),
			app.Fullscreen,
		)
		game := &ui.UI{}
		if err := game.Start(w); err != nil {
			fmt.Println(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
