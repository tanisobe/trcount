package trmon

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

func setKeybindgings(g *gocui.Gui, mw *MainWidget, nw *NarrowWidget) {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", 'k', gocui.ModNone, upCursor); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", 'j', gocui.ModNone, downCursor); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", '/', gocui.ModNone, changeRegexp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", gocui.KeyCtrlD, gocui.ModNone, pageDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", gocui.KeyCtrlU, gocui.ModNone, pageUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", 'd', gocui.ModNone, toggleDownIF(mw)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", 'u', gocui.ModNone, toggleUnit(mw)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", 'p', gocui.ModNone, togglebps(mw)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", 'h', gocui.ModNone, createHelp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("help", 'q', gocui.ModNone, terminateHelp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("help", 'h', gocui.ModNone, terminateHelp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("regexp", gocui.KeyEnter, gocui.ModNone, (nallowRegexp(nw))); err != nil {
		log.Panicln(err)
	}
}

// handler

//quit end app
func quit(*gocui.Gui, *gocui.View) error {
	return gocui.ErrQuit
}

//downCursor move down cusor
func downCursor(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

//upCursor move down cusor
func upCursor(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func pageDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		_, vy := v.Size()
		if err := v.SetCursor(cx, cy+vy); err != nil {
			if err := v.SetOrigin(ox, oy+vy); err != nil {
				return err
			}
		}

	}
	return nil
}

func pageUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		_, vy := v.Size()
		if err := v.SetCursor(cx, cy-vy); err != nil {
			if oy > vy-1 {
				if err := v.SetOrigin(ox, oy-vy); err != nil {
					return err
				}
			} else {
				v.SetCursor(cx, 0)
				v.SetOrigin(ox, 0)
			}
		}
	}
	return nil
}

func changeRegexp(g *gocui.Gui, v *gocui.View) error {
	if _, err := g.SetCurrentView("regexp"); err != nil {
		return err
	}
	return nil
}

func nallowRegexp(nw *NarrowWidget) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		expr, _ := v.Line(0)
		nw.setRegexp(expr)
		if _, err := g.SetCurrentView("main"); err != nil {
			return err
		}
		return nil
	}
}

func createHelp(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	v, err := g.SetView("help", maxX/10, maxY/5, maxX*8/10, maxY*5/6)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, helpMessage)
	}
	if _, err := g.SetCurrentView("help"); err != nil {
		return err
	}
	return nil
}

func terminateHelp(g *gocui.Gui, v *gocui.View) error {
	if _, err := g.SetCurrentView("main"); err != nil {
		return err
	}
	if err := g.DeleteView("help"); err != nil {
		return err
	}
	return nil
}

func toggleDownIF(mw *MainWidget) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if mw.displayDownIF {
			mw.displayDownIF = false
		} else {
			mw.displayDownIF = true
		}
		return nil
	}
}

func toggleUnit(m *MainWidget) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		switch m.unit {
		case Bps:
			m.setUnit(Kbps)
		case Kbps:
			m.setUnit(Mbps)
		case Mbps:
			m.setUnit(Bps)
		case Pps:
			m.setUnit(Kpps)
		case Kpps:
			m.setUnit(Mpps)
		case Mpps:
			m.setUnit(Pps)
		default:
			return fmt.Errorf("Unspecified value %v", m.unit)
		}
		return nil
	}
}

func togglebps(m *MainWidget) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if m.displaybps {
			m.displaybps = false
			return m.setUnit(Pps)
		}
		m.displaybps = true
		return m.setUnit(Bps)
	}
}
