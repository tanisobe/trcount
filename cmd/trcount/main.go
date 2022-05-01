package main

import (
	"flag"
	"fmt"

	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/tanisobe/trcount/snmp"
)

var (
	revision string
	dbg_mode *bool
)

func main() {
	snmp.Expr = flag.String("e", "", "Narrow down to IFs that match with regular expressions. Matching target is IF name and IF Description")
	dbg_mode = flag.Bool("debug", false, "start with debug mode. deubg mode dump trace log")
	cstring := flag.String("c", "public", "snmp community string.")
	interval := flag.Int("i", 5, "SNMP polling interval [sec]. minimum 5")
	flag.Parse()
	var dlog *log.Logger
	if *dbg_mode {
		wr, err := os.Create("trcount" + strconv.FormatInt(time.Now().Unix(), 10) + ".log")
		defer wr.Close()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		dlog = log.New(wr, "", log.LstdFlags|log.LUTC)
	} else {
		dlog = log.New(ioutil.Discard, "", log.LstdFlags|log.LUTC)
	}

	if *interval < 5 {
		log.Println("Too short interval, The minimum SNMP polling interval is 5 seconds")
		os.Exit(1)
	}
	run(dlog, cstring, *interval, flag.Args())
}

func run(dlog *log.Logger, cstring *string, interval int, args []string) {
	dlog.Println("start with debug mode")

	hosts := make([]*snmp.Host, 0)
	for _, h := range args {
		hosts = append(hosts, snmp.NewHost(h, snmp.DLogger(dlog), snmp.Cstring(cstring)))
	}

	for _, h := range hosts {
		dlog.Printf("Initalize %v", h.Name)
		if err := h.InitHost(); err != nil {
			log.Printf("Error: Iitialize %v, %v\n", h, err)
		}
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	// Update host information every unit time
	for _, host := range hosts {
		h := host
		go func() {
			for {
				h.Update()
				g.Update(func(g *gocui.Gui) error { return nil })
				time.Sleep(time.Duration(interval) * time.Second)
			}
		}()
	}

	g.Cursor = true
	g.Highlight = true
	mw := &snmp.MainWidget{Name: "main", Hosts: hosts}
	rw := &snmp.RegexpWidget{Name: "regexp"}

	g.SetManager(mw, rw)
	snmp.SetKeybindgings(g)

	maxX, maxY := g.Size()
	v, _ := g.SetView("main", 0, 0, maxX-2, maxY-3)
	g.SetCurrentView("main")
	v.SetCursor(0, 0)
	v, err = g.SetView("regexp", 0, maxY-2, maxX-2, maxY)
	fmt.Fprintf(v, *snmp.Expr)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
