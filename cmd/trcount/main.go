package main

import (
	"bytes"
	"flag"
	"fmt"

	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"

	"text/tabwriter"

	"github.com/jroimartin/gocui"
	"github.com/tanisobe/trcount/snmp"
)

var (
	revision      string
	dbg_mode      *bool
	regexp_narrow *string
)

type MainWidget struct {
	name  string
	hosts []*snmp.Host
}

type RegexpWidget struct {
	name string
}

func (m *MainWidget) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	v, err := g.SetView(m.name, 0, 0, maxX-2, maxY-3)
	if err != nil {
		if err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
		}
	}
	v.Highlight = true
	v.SelFgColor = gocui.ColorBlack
	v.SelBgColor = gocui.ColorGreen
	v.Clear()
	m.print(v)
	return nil
}

func (m *MainWidget) print(v *gocui.View) {
	tw := tabwriter.NewWriter(v, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)

	var keys []int
	for k := range m.hosts {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	narrowed := bytes.NewBufferString("")
	other := bytes.NewBufferString("")

	fmt.Fprintln(tw, "Name\tIF\tStatus\tIN[kb/s]\tOUT[kb/s]\tIn Error\tOut Error\tDesc\t")
	for _, k := range keys {
		m.printHost(narrowed, other, m.hosts[k])
	}
	fmt.Fprintf(tw, narrowed.String())
	fmt.Fprintf(tw, other.String())
	tw.Flush()
}

func (m *MainWidget) printHost(narrowed *bytes.Buffer, other *bytes.Buffer, h *snmp.Host) {
	var keys []int
	for k := range h.IFs {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	r, err := regexp.Compile(*regexp_narrow)
	if err != nil {
		fmt.Println(err)
	}

	for _, k := range keys {
		s := fmt.Sprintf("%v %v", h.IFs[k].Desc, h.IFs[k].Alias)
		if r.MatchString(s) {
			fmt.Fprintf(narrowed, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t\n", h.Name, h.IFs[k].Desc, h.IFs[k].OperStatus, h.IFs[k].InOctets.FlowRate()*8/1024, h.IFs[k].OutOctets.FlowRate()*8/1024, h.IFs[k].InError.Last, h.IFs[k].OutError.Last, h.IFs[k].Alias)
		} else {
			fmt.Fprintf(other, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t\n", h.Name, h.IFs[k].Desc, h.IFs[k].OperStatus, h.IFs[k].InOctets.FlowRate()*8/1024, h.IFs[k].OutOctets.FlowRate()*8/1024, h.IFs[k].InError.Last, h.IFs[k].OutError.Last, h.IFs[k].Alias)
		}
	}
}

func (w *RegexpWidget) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	v, err := g.SetView(w.name, 0, maxY-2, maxX-2, maxY)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	v.Editable = true
	v.FgColor = gocui.ColorWhite
	v.Editor = &Editor{}
	return nil
}

type Editor struct{}

func (e *Editor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	}
}

func main() {
	regexp_narrow = flag.String("e", "", "Narrow down to IFs that match with regular expressions. Matching target is IF name and IF Description")
	dbg_mode = flag.Bool("debug", false, "start with debug mode. deubg mode dump trace log")
	cstring := flag.String("c", "public", "snmp community string.")
	flag.Parse()
	var dlog *log.Logger
	if *dbg_mode {
		wr, err := os.Create("trcount" + strconv.FormatInt(time.Now().Unix(), 10) + ".log")
		defer wr.Close()
		if err != nil {
			log.Panicln(err)
		}
		dlog = log.New(wr, "", log.LstdFlags|log.LUTC)
	} else {
		dlog = log.New(ioutil.Discard, "", log.LstdFlags|log.LUTC)
	}

	run(dlog, cstring, flag.Args())
}

func run(dlog *log.Logger, cstring *string, args []string) {
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

	// Update host information every unit time
	for _, host := range hosts {
		h := host
		go func() {
			for {
				h.Update()
				time.Sleep(5 * time.Second)
			}
		}()
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true
	g.Highlight = true
	mw := &MainWidget{"main", hosts}
	rw := &RegexpWidget{"regexp"}

	g.SetManager(mw, rw)
	setKeybindgings(g)

	// Update View every seconds
	go func() {
		for {
			time.Sleep(1 * time.Second)
			g.Update(func(g *gocui.Gui) error { return nil })
		}
	}()

	maxX, maxY := g.Size()
	v, _ := g.SetView("main", 0, 0, maxX-2, maxY-3)
	g.SetCurrentView("main")
	v.SetCursor(0, 0)
	v, err = g.SetView("regexp", 0, maxY-2, maxX-2, maxY)
	fmt.Fprintf(v, *regexp_narrow)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func setKeybindgings(g *gocui.Gui) {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", 'k', gocui.ModNone, upCursor); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", 'j', gocui.ModNone, downCursor); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", '/', gocui.ModNone, changeRegrep); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("regexp", gocui.KeyEnter, gocui.ModNone, nallowRegrep); err != nil {
		log.Panicln(err)
	}
}

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

func changeRegrep(g *gocui.Gui, v *gocui.View) error {
	if _, err := g.SetCurrentView("regexp"); err != nil {
		return err
	}
	return nil
}

func nallowRegrep(g *gocui.Gui, v *gocui.View) error {
	*regexp_narrow, _ = v.Line(0)
	if _, err := g.SetCurrentView("main"); err != nil {
		return err
	}
	return nil
}
