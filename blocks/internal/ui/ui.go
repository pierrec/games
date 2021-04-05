package ui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	. "golang.org/x/image/colornames"

	"github.com/pierrec/games/blocks/internal/widgets"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type blockID,gameState -linecomment -output ui_string.go

type uiState uint8

const (
	uiNone uiState = iota
	uiHome
	uiScores
	uiGame
	uiGameOver
	uiSettings
	uiQuit
)

type UI struct {
	Config   string // file name
	theme    theme
	state    uiState
	home     home
	scores   scoreboard
	game     game
	settings settings
}

type config struct {
	Level        int           `json:"Level"`
	Keys         []keymapEntry `json:"Keys"`
	BlockColor   texture       `json:"blockcolor"`
	BlockPattern texture       `json:"blockpattern"`
	Scores       []scoreEntry  `json:"scores"`
}

type themeArea struct { // app areas
	Background color.NRGBA
	Foreground color.NRGBA
	Border     widget.Border
	Padding    unit.Value
}

type theme struct {
	Background color.NRGBA // app background
	Text       struct {
		Color  color.NRGBA
		Size   unit.Value
		Shaper text.Shaper
		Font   text.Font
	}
	Area themeArea
	Game themeArea
}

func (ui *UI) Start(w *app.Window) (err error) {
	defer func() {
		er := ui.saveConfig()
		if err == nil {
			err = er
		}
	}()
	shaper := text.NewCache(fontCollection[:])
	th := theme{
		Text: struct {
			Color  color.NRGBA
			Size   unit.Value
			Shaper text.Shaper
			Font   text.Font
		}{
			Color:  color.NRGBA(Gold),
			Size:   unit.Sp(12),
			Shaper: shaper,
		},
		Background: color.NRGBA(Black),
		Area: themeArea{
			Background: color.NRGBA(Gainsboro),
			Foreground: color.NRGBA(Gold),
			Border: widget.Border{
				Color:        color.NRGBA(Blue),
				CornerRadius: unit.Dp(10),
				Width:        unit.Dp(2),
			},
			Padding: unit.Dp(6),
		},
		Game: themeArea{
			Background: color.NRGBA(Black),
			Foreground: color.NRGBA(White),
			Border: widget.Border{
				Color:        color.NRGBA(Blue),
				CornerRadius: unit.Dp(10),
				Width:        unit.Dp(2),
			},
			Padding: unit.Dp(6),
		},
	}

	*ui = UI{
		theme:  th,
		Config: "blocks.cfg",
	}

	ops := new(op.Ops)
	evs := w.Events()
	for ui.state != uiQuit {
		select {
		case ev := <-evs:
			switch e := ev.(type) {
			case system.DestroyEvent:
				return e.Err

			case system.FrameEvent:
				gtx := layout.NewContext(ops, e)
				ui.Layout(gtx)
				e.Frame(gtx.Ops)
			}
		case <-ui.game.Tick():
			ui.game.Update(0)
			w.Invalidate()
		}
	}
	return nil
}

func (ui *UI) saveConfig() (err error) {
	dir, err := app.DataDir()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = fmt.Errorf("config in %s: %w", dir, err)
		}
	}()
	fName := filepath.Join(dir, ui.Config)
	f, err := os.OpenFile(fName, os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		er := f.Close()
		if err == nil {
			err = er
		}
	}()
	if err = f.Truncate(0); err != nil {
		return
	}
	var cfg config
	for _, v := range []interface{ saveConfig(*config) }{
		&ui.settings, &ui.home, &ui.scores,
	} {
		v.saveConfig(&cfg)
	}
	bts, err := json.Marshal(cfg)
	if err != nil {
		return
	}
	_, err = f.Write(bts)
	return
}

func (ui *UI) loadConfig() (err error) {
	dir, err := app.DataDir()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = fmt.Errorf("config in %s: %w", dir, err)
		}
	}()
	fName := filepath.Join(dir, ui.Config)
	f, err := os.OpenFile(fName, os.O_RDONLY, 0644)
	var cfg config
	switch {
	case err == nil:
		defer f.Close()
		bts, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(bts, &cfg); err != nil {
			return err
		}
	case os.IsNotExist(err):
	default:
		return err
	}
	for _, v := range []interface{ loadConfig(*config) }{
		&ui.settings, &ui.home, &ui.scores,
	} {
		v.loadConfig(&cfg)
	}
	return nil
}

func (ui *UI) init() {
	if ui.state != uiNone {
		return
	}
	ui.state = uiHome
	label := widgets.Label{
		Color:  color.NRGBA(Gold),
		Shaper: ui.theme.Text.Shaper,
		Size:   ui.theme.Text.Size,
		Font:   ui.theme.Text.Font,
		Inset: layout.Inset{
			Top:    unit.Dp(6),
			Bottom: unit.Dp(6),
		},
	}
	menu := widgets.Menu{
		List:   layout.List{Axis: layout.Vertical},
		Label:  label,
		Hover:  label.Color,
		Border: ui.theme.Area.Border,
	}
	ui.home = home{
		Menu: menu,
		Title: title{
			Background: ui.theme.Background,
		},
		Version: widgets.Label{
			Color:  color.NRGBA(Tan),
			Shaper: ui.theme.Text.Shaper,
			Size:   unit.Sp(12),
			Font:   ui.theme.Text.Font,
			Inset:  layout.UniformInset(unit.Dp(3)),
		},
	}
	ui.scores = scoreboard{
		Menu:    menu,
		Padding: ui.theme.Area.Padding,
	}
	ui.settings = settings{
		Menu:            menu,
		Padding:         ui.theme.Area.Padding,
		SelectedBg:      white,
		SelectedFg:      black,
		SelectedColor:   blackT,
		SelectedPattern: cornerT,
	}
	ui.game = game{
		Menu:       menu,
		ScoreLabel: label,
		Label:      label,
		Background: ui.theme.Game.Background,
		Border:     ui.theme.Game.Border.Color,
		Padding:    unit.Dp(6),
		KeyMap:     ui.settings.Key,
	}

	if err := ui.loadConfig(); err != nil {
		ui.home.Error = err
	}
}

func (ui *UI) update() {
	switch ui.state {
	case uiHome:
		level := ui.home.Level()
		ui.game.StartLevel = level
		ui.home.Title.Gravity = gameGravity(level)
		ui.home.Title.Texture = ui.settings.Texture()
		switch i := ui.home.Menu.Clicked(); i {
		case homeStartGame:
			ui.state = uiGame
			ui.game.BlockTexture = ui.settings.Texture()
			ui.game.Start()
		case homeScoreBoard:
			ui.state = uiScores
		case homeSettings:
			ui.state = uiSettings
		case homeQuitGame:
			ui.state = uiQuit
		}
	case uiScores:
		switch i := ui.scores.Menu.Clicked(); i {
		case scoreboardBack:
			ui.state = uiHome
		}
	case uiGame:
		switch ui.game.state {
		case gameOver:
			ui.state = uiGameOver
		case gameLeft:
			ui.state = uiHome
		}
	case uiGameOver:
		if score, over := ui.game.Over(); over {
			ui.state = uiHome
			if ui.scores.NewScore(score) {
				ui.state = uiScores
			}
		}
	case uiSettings:
		switch ui.settings.Menu.Clicked() {
		case settingsBack:
			ui.state = uiHome
			ui.home.Error = ui.saveConfig()
		}
	}
}

func (ui *UI) Layout(gtx layout.Context) layout.Dimensions {
	ui.init()
	ui.update()

	// Background color.
	paint.FillShape(gtx.Ops, ui.theme.Background, clip.Rect{Max: gtx.Constraints.Max}.Op())

	gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, 1500)
	switch ui.state {
	case uiHome:
		return ui.home.Layout(gtx)
	case uiScores:
		return ui.scores.Layout(gtx)
	case uiGame, uiGameOver:
		return ui.game.Layout(gtx)
	case uiSettings:
		return ui.settings.Layout(gtx)
	}
	return layout.Dimensions{}
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
