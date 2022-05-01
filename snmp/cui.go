package snmp

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"sort"

	"github.com/jroimartin/gocui"
	"github.com/olekukonko/tablewriter"
)

var Expr *string

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
	v.Highlight = true
	v.SelFgColor = gocui.ColorBlack
	v.SelBgColor = gocui.ColorGreen
	v.Clear()
	m.print(v)
	return nil
}

func (m *MainWidget) print(v *gocui.View) {
	t := tablewriter.NewWriter(v)
	t.SetRowLine(false)
	t.SetBorder(false)
	t.SetAutoFormatHeaders(false)
	t.SetHeader([]string{"Name", "I/F", "Status", "IN[kb/s]", "OUT[kb/s]", "InErr", "OutErr", "Description"})
	t.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
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
			{tablewriter.Normal, tablewriter.FgCyanColor},
			{tablewriter.Normal, tablewriter.FgCyanColor},
			{tablewriter.Normal, tablewriter.FgCyanColor},
			{tablewriter.Normal, tablewriter.FgCyanColor},
			{tablewriter.Normal, tablewriter.FgCyanColor},
			{tablewriter.Normal, tablewriter.FgCyanColor},
			{tablewriter.Normal, tablewriter.FgCyanColor},
			{tablewriter.Normal, tablewriter.FgCyanColor},
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
	r, err := regexp.Compile(*Expr)
	if err != nil {
		fmt.Println(err)
	}

	for _, k := range keys {
		s := fmt.Sprintf("%v %v", h.IFs[k].Desc, h.IFs[k].Alias)
		data := []string{
			h.Name,
			h.IFs[k].Desc,
			h.IFs[k].OperStatus,
			fmt.Sprint(h.IFs[k].InOctets.FlowRate() * 8 / 1024),
			fmt.Sprint(h.IFs[k].OutOctets.FlowRate() * 8 / 1024),
			fmt.Sprint(h.IFs[k].InError.Last),
			fmt.Sprint(h.IFs[k].OutError.Last),
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
	v.Editor = &Editor{*bytes.NewBufferString(*Expr)}
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
	*Expr, _ = v.Line(0)
	if _, err := g.SetCurrentView("main"); err != nil {
		return err
	}
	return nil
}
