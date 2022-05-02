package snmp

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/jroimartin/gocui"
	"github.com/olekukonko/tablewriter"
)

type UnitCalc func(int64) int64

var (
	Expr          string              = ""
	DisplayDownIF bool                = true
	unit          string              = "kbps"
	unitCalc      func(x int64) int64 = func(x int64) int64 { return x * 8 / 1024 }
)

const (
	helpMessage = `
	h: help
	j: down cursor
	k: up cursor
	u: toggle the unit of display traffic [bps][kbps][mbps]
	d: toggle the display of Down I/F
	/: narrow down with regex
	   Targets of narrowing down are Description and I/F
	q: quit
	`
)

type MainWidget struct {
	Name  string
	Hosts []*Host
}

type RegexpWidget struct {
	Name string
}

type Editor struct {
	Buffer bytes.Buffer
}

func (m *MainWidget) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	v, err := g.SetView(m.Name, 0, 0, maxX-2, maxY-3)
	if err != nil {
		if err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
		}
	}
	v.Clear()
	m.print(v)
	return nil
}

func (m *MainWidget) print(v *gocui.View) {
	t := tablewriter.NewWriter(v)
	t.SetRowLine(false)
	t.SetBorder(false)
	t.SetAutoWrapText(false)
	t.SetAutoFormatHeaders(false)
	t.SetHeader([]string{
		"Name",
		"I/F",
		"Status",
		fmt.Sprintf("IN[%v]", unit),
		fmt.Sprintf("OUT[%v]", unit),
		"InErr",
		"OutErr",
		"InDis",
		"OutDis",
		"Description",
	})
	t.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor, tablewriter.FgBlackColor},
	)

	var keys []int
	for k := range m.Hosts {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	narrowed := make([][]string, 0, 300)
	for col := range narrowed {
		narrowed[col] = make([]string, 0, 10)
	}
	other := make([][]string, 0, 300)
	for col := range other {
		other[col] = make([]string, 0, 10)
	}

	for _, k := range keys {
		m.printHost(&narrowed, &other, m.Hosts[k])
	}
	for _, row := range narrowed {
		t.Rich(row, []tablewriter.Colors{
			{tablewriter.FgCyanColor},
			{tablewriter.FgCyanColor},
			{tablewriter.FgCyanColor},
			{tablewriter.FgCyanColor},
			{tablewriter.FgCyanColor},
			{tablewriter.FgCyanColor},
			{tablewriter.FgCyanColor},
			{tablewriter.FgCyanColor},
			{tablewriter.FgCyanColor},
			{tablewriter.FgCyanColor},
		})
	}
	for _, row := range other {
		t.Append(row)
	}
	t.Render()
}

func (m *MainWidget) printHost(narrowed *[][]string, other *[][]string, h *Host) {
	var keys []int
	for k := range h.IFs {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	r, err := regexp.Compile(Expr)
	if err != nil {
		fmt.Println(err)
	}

	for _, k := range keys {
		if !DisplayDownIF && h.IFs[k].OperStatus == "Down" {
			continue
		}
		s := fmt.Sprintf("%v %v", h.IFs[k].Desc, h.IFs[k].Alias)
		data := []string{
			h.Name,
			h.IFs[k].Desc,
			h.IFs[k].OperStatus,
			humanize.Comma(unitCalc(h.IFs[k].InOctets.Rate)),
			humanize.Comma(unitCalc(h.IFs[k].OutOctets.Rate)),
			humanize.Comma(h.IFs[k].InError.Diff),
			humanize.Comma(h.IFs[k].OutError.Diff),
			humanize.Comma(h.IFs[k].InDiscards.Diff),
			humanize.Comma(h.IFs[k].OutDiscards.Diff),
			h.IFs[k].Alias,
		}
		if r.MatchString(s) {
			*narrowed = append(*narrowed, data)
		} else {
			*other = append(*other, data)
		}
	}
}

func (w *RegexpWidget) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	v, err := g.SetView(w.Name, 0, maxY-2, maxX-2, maxY)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	v.Editable = true
	v.FgColor = gocui.ColorWhite
	v.Editor = &Editor{*bytes.NewBufferString(Expr)}
	return nil
}

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

func SetKeybindgings(g *gocui.Gui) {
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
	if err := g.SetKeybinding("main", 'd', gocui.ModNone, toggleDisplayDownIF); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", 'u', gocui.ModNone, toggleUnit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("main", 'h', gocui.ModNone, viewHelp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("help", 'q', gocui.ModNone, finishHelp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("help", 'h', gocui.ModNone, finishHelp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("regexp", gocui.KeyEnter, gocui.ModNone, nallowRegexp); err != nil {
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

func changeRegexp(g *gocui.Gui, v *gocui.View) error {
	if _, err := g.SetCurrentView("regexp"); err != nil {
		return err
	}
	return nil
}

func nallowRegexp(g *gocui.Gui, v *gocui.View) error {
	Expr, _ = v.Line(0)
	if _, err := g.SetCurrentView("main"); err != nil {
		return err
	}
	return nil
}

func toggleDisplayDownIF(g *gocui.Gui, v *gocui.View) error {
	if DisplayDownIF {
		DisplayDownIF = false
	} else {
		DisplayDownIF = true
	}
	return nil
}

func toggleUnit(g *gocui.Gui, v *gocui.View) error {
	switch unit {
	case "bps":
		unit = "kbps"
		unitCalc = func(x int64) int64 { return x * 8 / 1024 }
	case "kbps":
		unit = "mbps"
		unitCalc = func(x int64) int64 { return x * 8 / 1024 / 1024 }
	case "mbps":
		unit = "bps"
		unitCalc = func(x int64) int64 { return x * 8 }
	}
	return nil
}

func viewHelp(g *gocui.Gui, v *gocui.View) error {
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

func finishHelp(g *gocui.Gui, v *gocui.View) error {
	if _, err := g.SetCurrentView("main"); err != nil {
		return err
	}
	if err := g.DeleteView("help"); err != nil {
		return err
	}
	return nil
}
