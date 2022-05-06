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

	// Update host information every unit time
	a.log.Debug().Msg("Start goroutin to update host")
	ctx, cancel := context.WithCancel(context.Background())
	ctx1 := context.Context(ctx)
	defer cancel()

	a.suicide(ctx, c.Lifespan)
	a.updateHosts(ctx1, c.Interval)

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
	v, err := a.gui.SetView("main", 0, 0, maxX-2, maxY-3)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			a.log.Error().Msgf("%v", err)
			return err
		}
	}

	a.gui.SetCurrentView("main")
	v.SetCursor(0, 0)
	if v, err = a.gui.SetView("regexp", 0, maxY-2, maxX-2, maxY); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			a.log.Error().Msgf("%v", err)
			return err
		}
	}
	fmt.Fprintf(v, expr)
	return nil
}

func (a *App) updateHosts(ctx context.Context, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for _, host := range a.hosts {
		h := host
		c := context.Context(ctx)
		go func(ctx context.Context, h *Host) {
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
	go func() {
		for {
			select {
			case <-after:
				a.log.Info().Msgf("It has been running for more than %v seconds, so it will end the process.", lifespan)
				a.gui.Update(func(g *gocui.Gui) error { return gocui.ErrQuit })
			case <-ctx.Done():
				a.log.Debug().Msg("Done suicide goroutine")
				return
			}
		}
	}()
}
