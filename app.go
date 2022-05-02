package trmon

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jroimartin/gocui"
)

func Run(community string, interval int, expr string, isDebug bool, f io.Writer, args []string) {
	log := NewLogger(isDebug, f)

	log.Debug().Msg("start with debug mode")

	hosts := make([]*Host, 0)
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
		os.Exit(0)
	}

	// CUI Initialize
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
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panic().Msgf("%v", err)
	}
}
