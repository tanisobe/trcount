package trmon

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/jroimartin/gocui"
)

func Run(community string, interval int, expr string, isDebug bool, f io.Writer, args []string) error {
	log := NewLogger(isDebug, f)
	hosts := make([]*Host, 0)

	// SNMP host Initalize
	log.Debug().Msg("SNMP host init")
	for _, name := range args {
		host, err := NewHost(name, community, log)
		if err != nil {
			log.Warn().Msgf("%v can't initalized : %v", name, err)
			continue
		}
		hosts = append(hosts, host)
	}

	if len(hosts) == 0 {
		log.Warn().Msgf("No accesstable host")
		return errors.New("No accesstable host")
	}
	// CUI Initialize
	log.Debug().Msg("CUI initalize")
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Error().Msgf("%v", err)
	}
	defer g.Close()
	g.Cursor = true
	g.Highlight = true
	nw := NewNarrowWidget("regexp", expr, log)
	mw := NewMainWidget("main", hosts, nw, log)
	g.SetManager(mw, nw)
	setKeybindgings(g, mw, nw)

	// View Initialize
	maxX, maxY := g.Size()
	v, err := g.SetView("main", 0, 0, maxX-2, maxY-3)
	if err != nil {
	}

	g.SetCurrentView("main")
	v.SetCursor(0, 0)
	v, err = g.SetView("regexp", 0, maxY-2, maxX-2, maxY)
	fmt.Fprintf(v, expr)

	// Update host information every unit time
	log.Debug().Msg("Start goroutin to update host")
	for _, host := range hosts {
		h := host
		go func(h *Host) {
			for {
				log.Debug().Msgf("Update %v", h.Name)
				h.Update()
				log.Debug().Msg("Update Display")
				g.Update(func(g *gocui.Gui) error { return nil })
				time.Sleep(time.Duration(interval) * time.Second)
			}
		}(h)
	}

	// mainloop for CUI Event
	log.Debug().Msg("Start CUI main loop")
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Error().Msgf("%v", err)
		return err
	}
	return nil
}
