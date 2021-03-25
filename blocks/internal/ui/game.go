package ui

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"git.sr.ht/~pierrec/giox/widgetx"

	"github.com/pierrec/games/blocks/internal/widgets"
)

type gameState uint8

const (
	gameNone      gameState = iota // NOSTATE
	gameRunning                    // RUNNING
	gameFullLines                  // FULLLINES
	gameLineAnim                   // LINEANIM
	gamePaused                     // PAUSED
	gameOver                       // GAME OVER
	gameLeft                       // GAMELEFT
)

const (
	gamePause = iota
	gameContinue
	gameBack
	game_
)

type lines struct {
	First int
	Last  int
	anim  widgets.Anim
}

func (l *lines) next(pos int) (int, time.Duration) {
	pos += 2 * (l.Last - l.First)
	d := math.Sqrt(float64(pos))
	return pos, time.Duration(d)
}

func (l *lines) Start(n int) {
	if l.anim.Next == nil {
		l.anim.Next = l.next
	}
	l.anim.Start(1, (l.Last-l.First)*n)
}

func (l *lines) Animate(gtx layout.Context) (pos int, done bool) {
	pos = l.anim.Animate(gtx)
	done = !l.anim.Animating()
	return
}

// game manages the game window with its board, score...
type game struct {
	ScoreLabel   widgets.Label
	Label        widgets.Label
	Menu         widgets.Menu
	Background   color.NRGBA
	Border       color.NRGBA
	Padding      unit.Value
	StartLevel   int
	KeyMap       func(string) int
	BlockTexture texture

	state    gameState
	overlay  widgetx.Modal
	ticker   *time.Ticker
	paused   *time.Ticker
	current  block
	area     grid
	lines    lines
	next     block
	areaNext grid
	score    score
}

// drawGridBorder draws an invisible border around the grid
// (except the top).
func (ui *game) drawGridBorder() {
	sz := ui.area.Size()
	cols, rows := sz.X, sz.Y
	t := invisibleT
	for x := 0; x < cols; x++ {
		ui.area.Set(x, rows-1, t) // bottom line
	}
	for y := 0; y < rows; y++ {
		ui.area.Set(0, y, t)      // left line
		ui.area.Set(cols-1, y, t) // right line
	}
}

func (ui *game) setGridCellSize(gtx layout.Context) {
	rows := ui.area.Size().Y - 1 - 1 // border + top hidden row
	pad := gtx.Metric.Px(ui.Padding) * 2
	size := image.Point{
		X: (gtx.Constraints.Max.Y - pad) / rows,
		Y: (gtx.Constraints.Max.Y - pad) / rows,
	}
	ui.area.SetCellSize(size)
}

// Start starts the game by initializing blocks and the ticker.
func (ui *game) Start() {
	state := ui.state
	ui.state = gameRunning
	switch state {
	case gamePaused:
		ui.unpause()
		return
	case gameOver, gameLeft:
		ui.area.Clear()
		ui.drawGridBorder()
	}
	ui.score = score{
		Label:        ui.ScoreLabel,
		Padding:      ui.Padding.Scale(2),
		Level:        ui.StartLevel,
		LineColor:    ui.Border,
		LineHeight:   unit.Dp(1),
		LineOverflow: unit.Dp(6),
	}
	ui.setGravity()
	ui.current.InitRandom(ui.BlockTexture)
	ui.next.InitRandom(ui.BlockTexture)
}

// Pause pauses the game, without stopping the ticker.
// Resume by calling Start.
func (ui *game) Pause() {
	if ui.state == gameRunning {
		ui.state = gamePaused
		ui.pause()
	}
}

func (ui *game) pause() {
	ui.paused = ui.ticker
	ui.ticker = nil
}

func (ui *game) unpause() {
	ui.ticker = ui.paused
	ui.paused = nil
}

// Stop marks the game as over and clears the ticker.
func (ui *game) Stop() {
	switch ui.state {
	case gamePaused:
		ui.unpause()
		fallthrough
	case gameRunning:
		ui.ticker.Stop()
		ui.ticker = nil
	}
	ui.state = gameOver
}

func (ui *game) Over() (score []scoreData, over bool) {
	if ui.state == gameOver && ui.overlay.Changed() {
		return ui.score.Scores(), true
	}
	return nil, false
}

func (ui *game) Tick() <-chan time.Time {
	if ui.ticker != nil {
		return ui.ticker.C
	}
	return nil
}

// Update manages the game loop and is triggered when a new loop
// is to be started (upon ticker firing or user movement blocked).
// It tries to move the current block down and if unsuccessful,
// use the next block as the current one and try again. If it fails,
// then the game is over, at which point the ticker is stopped and cleared.
func (ui *game) Update(softDrop int) {
	if ui.state != gameRunning {
		return
	}
	if ui.current.MoveDown(&ui.area) {
		// The current block successfully moved down.
		return
	}
	// The current block can no longer move.
	ui.checkFullLines()
	ui.score.NewBlock(softDrop)
	// Use a new block.
	ui.current = ui.next
	ui.next.InitRandom(ui.BlockTexture)
}

func (ui *game) checkFullLines() {
	// Detect full lines.
	area := ui.area.Slice(
		image.Pt(0, 1), // hide the first line
		ui.area.Size(),
	)

	pos := ui.current.Pos()
	start, end := -1, -1                  // indexes of the first and last full lines.
	sz := area.Size().Sub(image.Pt(1, 1)) // with left and bottom wall correction
	xn := sz.X
	_, h := ui.current.Dims()
	yn := min(pos.Y+h, sz.Y)
fullLoop:
	for y := pos.Y; y < yn; y++ {
		for x := 1; x < xn; x++ {
			if area.Get(x, y) == transparentT {
				continue fullLoop
			}
		}
		// Full line!
		if start < 0 {
			start = y
		}
		end = y + 1
	}
	if start >= 0 {
		ui.state = gameFullLines
		ui.lines.First = start
		ui.lines.Last = end
	}
}

func (ui *game) setGravity() {
	level := ui.score.CurrentLevel()
	d := gameGravity(level)
	if ui.ticker == nil {
		ui.ticker = time.NewTicker(d)
	} else {
		ui.ticker.Reset(d)
	}
}

type queue []event.Event

func (q queue) Events(event.Tag) []event.Event { return q }

func (ui *game) init(gtx layout.Context) {
	if ui.area.Size() == (image.Point{}) {
		bg := ui.Background
		bg.A = 128
		ui.overlay = widgetx.Modal{
			Background: bg,
			Keys: []string{key.NameEscape,
				key.NameEnter, key.NameReturn,
				key.NameSpace},
		}
		// Grid with a border
		// and the first line as hidden to allow rotation while on top.
		cols, rows := 10+2, 20+1+1
		ui.area.Init(cols, rows)
		ui.area.Background = ui.Background
		ui.areaNext.Background = ui.Background
		ui.setGridCellSize(gtx)
		ui.drawGridBorder()
		ui.current.KeyMap = ui.KeyMap
		ui.next.KeyMap = ui.KeyMap
		ui.score.AnimBg = ui.Background
	}
}

func (ui *game) update(gtx layout.Context, evs []event.Event) {
	switch ui.state {
	case gameRunning:
		// Hide the pointer while playing unless it moves or is clicked.
		pointer.InputOp{
			Tag:   ui,
			Types: pointer.Move | pointer.Press,
		}.Add(gtx.Ops)
		//TODO on Linux, the cursor is kept hidden when moving out of the app window
		ptr := pointer.CursorNone
		for _, ev := range evs {
			switch e := ev.(type) {
			case key.Event:
				if e.State == key.Release {
					switch ui.KeyMap(e.Name) {
					case pauseGame:
						ui.Pause()
					}
				}
			case pointer.Event:
				ptr = pointer.CursorDefault
			}
		}
		pointer.CursorNameOp{Name: ptr}.Add(gtx.Ops)
	case gameFullLines:
		ui.state = gameLineAnim
		ui.lines.Start(ui.area.CellSize().Y)
		ui.pause()
	case gamePaused:
		switch ui.Menu.Clicked() {
		case gameContinue:
			ui.Start()
			op.InvalidateOp{}.Add(gtx.Ops)
		case gameBack:
			ui.state = gameLeft
			op.InvalidateOp{}.Add(gtx.Ops)
		}
	}
}

func (ui *game) Layout(gtx layout.Context) layout.Dimensions {
	ui.init(gtx)
	ui.setGridCellSize(gtx) // support window resizing

	// Background color.
	paint.FillShape(gtx.Ops, ui.Background, clip.Rect{Max: gtx.Constraints.Max}.Op())

	var showOverlay bool
	switch ui.state {
	case gamePaused, gameOver:
		showOverlay = true
		ui.update(gtx, nil)
	case gameFullLines, gameLineAnim:
		// Do not enable keyboard/mouse input while animating.
		ui.update(gtx, nil)
	case gameLeft:
		return layout.Dimensions{}
	default:
		//pointer.InputOp{Tag: ui}.Add(gtx.Ops)
		// Get all key events here and make them available to downstream widgets.
		// This allows catching the key to pause the game.
		key.InputOp{Tag: ui}.Add(gtx.Ops)
		key.FocusOp{Tag: ui}.Add(gtx.Ops)
		evs := gtx.Queue.Events(ui)
		gtx.Queue = queue(evs)
		ui.update(gtx, evs)
		// Display the current block.
		ui.current.Layout(gtx, &ui.area, ui.Update, ui.Stop)
	}

	var gridDims layout.Dimensions
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Horizontal,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				start := image.Pt(0, 1) // hide the first line
				end := ui.area.Size()
				area := ui.area.Slice(start, end)
				gridDims = ui.layoutPanel(gtx, area.Layout)
				ui.animate(gtx)
				// Display the pause/game over overlays on top of the grid.
				if !showOverlay {
					return gridDims
				}
				gtx.Constraints = layout.Exact(gridDims.Size)
				ui.layoutOverlay(gtx)
				return gridDims
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gridDims.Size.X)
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return ui.layoutPanel(gtx, ui.layoutScore)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return ui.layoutPanel(gtx, ui.layoutNextBlock)
					}),
				)
			}),
		)
	})
}

func (ui *game) animate(gtx layout.Context) {
	if ui.state != gameLineAnim {
		return
	}
	start := ui.lines.First
	end := ui.lines.Last
	num := end - start
	if pos, done := ui.lines.Animate(gtx); !done {
		cell := ui.area.CellSize()
		xn := ui.area.Size().X
		pad := gtx.Metric.Px(ui.Padding)
		pt := image.Pt(pad, pad)
		y := cell.Y * end
		r := clip.Rect{
			Min: pt.Add(image.Pt(0, y-pos)),
			Max: pt.Add(image.Pt(cell.X*(xn-2), y)),
		}
		paint.FillShape(gtx.Ops, ui.Background, r.Op())
		return
	}
	ui.state = gameRunning
	ui.unpause()
	// Move down all non empty lines before start by end-start amount.
	xn := ui.area.Size().X - 1 // left wall correction applied
	for y := start; y+num >= 0; y-- {
		for x := 1; x < xn; x++ {
			t := transparentT
			if y >= 0 {
				t = ui.area.Get(x, y)
			}
			ui.area.Set(x, y+num, t)
		}
	}
	// Update the score.
	if ui.score.NewLines(num) {
		// Level changed: increase the gravity.
		ui.setGravity()
	}
	op.InvalidateOp{}.Add(gtx.Ops)
}

func (ui *game) layoutPanel(gtx layout.Context, panel layout.Widget) layout.Dimensions {
	return widget.Border{
		Color:        ui.Border,
		CornerRadius: ui.Padding,
		Width:        ui.Padding.Scale(0.5),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(ui.Padding).Layout(gtx, panel)
	})
}

func (ui *game) layoutOverlay(gtx layout.Context) {
	ui.overlay.Layout(gtx)
	pointer.InputOp{Tag: ui}.Add(gtx.Ops)

	macro := op.Record(gtx.Ops)
	n := game_
	if ui.state == gameOver {
		n = 1
	}
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X /= 2
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return ui.Menu.Layout(gtx, n, func(gtx layout.Context, i int) widgets.MenuItem {
			l := ui.Label
			var txt string
			noTitle := widgets.MenuNoTitle(func(gtx layout.Context) layout.Dimensions {
				inset := layout.Inset{
					Top:    unit.Dp(16),
					Right:  unit.Dp(24),
					Bottom: unit.Dp(16),
					Left:   unit.Dp(24),
				}
				return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return l.Layout(gtx, txt)
				})
			})
			switch ui.state {
			case gameOver:
				l.Font.Weight = text.Bold
				txt = ui.state.String()
				return noTitle
			case gamePaused:
				switch i {
				case gamePause:
					l.Font.Weight = text.Bold
					txt = ui.state.String()
					return noTitle
				case gameContinue:
					txt = "Continue"
				case gameBack:
					txt = "Quit"
				}
				return widgets.MenuButton(func(gtx layout.Context) layout.Dimensions {
					return l.Layout(gtx, txt)
				})
			}
			return widgets.MenuNoTitle(func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{}
			})
		})
	})
	op.Defer(gtx.Ops, macro.Stop())
}

func (ui *game) layoutScore(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, ui.score.Layout)
}

func (ui *game) layoutNextBlock(gtx layout.Context) layout.Dimensions {
	b := &ui.next
	g := &ui.areaNext
	g.Resize(b.Dims())
	g.Clear()
	g.SetCellSize(ui.area.CellSize())
	b.layout(g, false)
	return layout.Center.Layout(gtx, g.Layout)
}
