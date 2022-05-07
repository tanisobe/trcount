package trmon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/jroimartin/gocui"
)

type App struct {
	hosts []*Host
	gui   *gocui.Gui
	log   *Logger
}

type Config struct {
	Interval  int
	Lifespan  int
	Community string
	Expr      string
	IsDebug   bool
	Output    io.Writer
}

func (a *App) Run(hostnames []string, c *Config) error {

	a.hosts = make([]*Host, 0)
	a.log = NewLogger(c.IsDebug, c.Output)

	// SNMP host Initalize
	a.log.Debug().Msg("SNMP host init")
	if err := a.initHosts(hostnames, c.Community); err != nil {
		return err
	}
	// CUI Initialize
	var err error
	a.log.Debug().Msg("CUI initalize")
	a.gui, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		a.log.Error().Msgf("%v", err)
		return err
	}
	defer a.gui.Close()

	if err := a.initCUI(c.Expr); err != nil {
		a.log.Error().Msgf("%v", err)
		return err
	}

	a.log.Debug().Msg("Start background goroutin")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.suicide(ctx, c.Lifespan)
	a.updateHosts(ctx, c.Interval)
	a.showInitView(ctx, c.Interval)

	// mainloop for CUI Event
	a.log.Debug().Msg("Start CUI main loop")
	if err := a.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		a.log.Error().Msgf("%v", err)
		return err
	}
	return nil
}

func (a *App) initHosts(hostnames []string, community string) error {
	// SNMP host Initalize
	a.log.Debug().Msg("SNMP host init")
	for _, name := range hostnames {
		host, err := NewHost(name, community, a.log)
		if err != nil {
			a.log.Warn().Msgf("%v can't initalized : %v", name, err)
			continue
		}
		a.hosts = append(a.hosts, host)
	}

	if len(a.hosts) == 0 {
		a.log.Warn().Msgf("No accesstable host")
		return errors.New("No accesstable host")
	}
	return nil
}

func (a *App) initCUI(expr string) error {
	a.gui.Cursor = true
	a.gui.Highlight = true
	nw := NewNarrowWidget("regexp", expr, a.log)
	mw := NewMainWidget("main", a.hosts, nw, a.log)
	a.gui.SetManager(mw, nw)
	setKeybindgings(a.gui, mw, nw)

	// View Initialize
	maxX, maxY := a.gui.Size()
	v, err := a.newView("main", 0, 0, maxX-2, maxY-3, "")
	if err != nil {
		return err
	}
	a.gui.SetCurrentView("main")
	v.SetCursor(0, 0)
	_, err1 := a.newView("regexp", 0, maxY-2, maxX-2, maxY, expr)
	if err1 != nil {
		return err
	}
	return nil
}

func (a *App) updateHosts(ctx context.Context, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for _, host := range a.hosts {
		h := host
		c := context.Context(ctx)
		go func(ctx context.Context, h *Host) {
			a.log.Debug().Msgf("First Update %v", h.Name)
			h.Update()
			for {
				select {
				case <-ticker.C:
					a.log.Debug().Msgf("Update %v", h.Name)
					h.Update()
					a.log.Debug().Msg("Update Display")
					a.gui.Update(func(g *gocui.Gui) error { return nil })
				case <-ctx.Done():
					a.log.Debug().Msgf("Done update %v goroutine", h.Name)
					return
				}
			}
		}(c, h)
	}
}

func (a *App) suicide(ctx context.Context, lifespan int) {
	after := time.After(time.Duration(lifespan) * time.Second)
	c := context.Context(ctx)
	go func() {
		for {
			select {
			case <-after:
				a.log.Info().Msgf("It has been running for more than %v seconds, so it will end the process.", lifespan)
				a.gui.Update(func(g *gocui.Gui) error { return gocui.ErrQuit })
				return
			case <-c.Done():
				a.log.Debug().Msg("Done suicide goroutine")
				return
			}
		}
	}()
}

func (a *App) showInitView(ctx context.Context, lifespan int) error {
	timeout := time.After(time.Duration(lifespan) * time.Second)
	ticker := time.NewTicker(time.Duration(1) * time.Second)
	c := context.Context(ctx)

	maxX, maxY := a.gui.Size()
	iv, err := a.newView("init", maxX/10, 0, maxX*8/10, maxY/6, "Now initializing")
	if err != nil {
		return err
	}
	_, err1 := a.newView("help", maxX/10, maxY/5, maxX*8/10, maxY*9/10, helpMessage)
	if err1 != nil {
		return err1
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				l, _ := iv.Line(0)
				iv.Clear()
				fmt.Fprintln(iv, l+".")
				a.gui.Update(func(g *gocui.Gui) error { return nil })
			case <-timeout:
				a.log.Debug().Msgf("Delete Initialize View")
				a.gui.DeleteView("init")
				a.gui.DeleteView("help")
				ticker.Stop()
				return
			case <-c.Done():
				ticker.Stop()
				return
			}
		}
	}()
	return nil
}

func (a *App) newView(name string, x0, y0, x1, y1 int, msg string) (*gocui.View, error) {
	v, err := a.gui.SetView(name, x0, y0, x1, y1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			a.log.Error().Msgf("%v", err)
			return nil, err
		}
	}
	fmt.Fprintln(v, msg)
	return v, nil
}
