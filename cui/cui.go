package cui

import (
	"context"
	"github.com/illatior/task-scheduler/core/metric"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"sync"
)

type cui struct {
	isFullscreen bool

	t terminalapi.Terminal
	c *container.Container

	screenMu      sync.RWMutex
	currentScreen int
	screens       []Screen
}

func NewCui(t terminalapi.Terminal, screens ...Screen) (ConsoleUserInterface, error) {
	c, err := container.New(
		t,
		container.ID(SCREEN_ID),
		container.Border(linestyle.Light),
		container.BorderTitle("Task-scheduler"),
	)
	if err != nil {
		return nil, err
	}

	ui := &cui{
		isFullscreen: false,
		c:            c,
		t:            t,

		screenMu:      sync.RWMutex{},
		currentScreen: 0,
		screens:       screens,
	}
	err = ui.changeMainScreen()
	if err != nil {
		return nil, err
	}

	return ui, nil
}

func (ui *cui) Run(ctx context.Context, done chan<- bool) error {
	defer func() {
		done <- true
	}()

	ctx, cancel := context.WithCancel(ctx)
	subs := func (k *terminalapi.Keyboard) {
		var err error
		switch k.Key {
		case 'Q', 'q', keyboard.KeyCtrlC:
			cancel()
		case 'A', 'a':
			err = ui.PreviousScreen()
		case 'D', 'd':
			err = ui.NextScreen()
		case 'F', 'f':
			err = ui.ChangeFullscreenState()
		default:
			return
		}

		if err != nil {
			panic(err)
		}
	}

	defer func() {
		ui.t.Close()
	}()
	return termdash.Run(ctx, ui.t, ui.c, termdash.KeyboardSubscriber(subs))
}

func (ui *cui) AcceptMetric(m *metric.Result) {
	return
}

func (ui *cui) ChangeFullscreenState() error {
	ui.isFullscreen = !ui.isFullscreen

	return ui.changeMainScreen()
}

func (ui *cui) NextScreen() error {
	ui.screenMu.Lock()
	defer ui.screenMu.Unlock()

	ui.currentScreen++
	if ui.currentScreen == len(ui.screens) {
		ui.currentScreen = 0
	}

	return ui.changeMainScreen()
}

func (ui *cui) PreviousScreen() error {
	ui.screenMu.Lock()
	defer ui.screenMu.Unlock()

	ui.currentScreen--
	if ui.currentScreen < 0 {
		ui.currentScreen = len(ui.screens) - 1
	}

	return ui.changeMainScreen()
}

func (ui *cui) changeMainScreen() error { // FIXME after exiting fullscreen mode main BorderTitle and BorderStyle continues to be as body's one
	currentScreen := ui.screens[ui.currentScreen]

	builder := grid.New()
	if ui.isFullscreen {
		addElem(currentScreen.GetBody(), builder)
	} else {
		addElem(currentScreen.GetHeader(), builder)
		addElem(currentScreen.GetBody(), builder)
		addElem(currentScreen.GetFooter(), builder)
	}

	opts, err := builder.Build()
	if err != nil {
		return err
	}

	return ui.c.Update(SCREEN_ID, opts...)
}

func (ui *cui) IsFullscreen() bool {
	return ui.isFullscreen
}

func addElem(e grid.Element, b *grid.Builder) {
	if e != nil {
		b.Add(e)
	}
}
