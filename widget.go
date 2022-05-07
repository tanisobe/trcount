package trmon

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/jroimartin/gocui"
	"github.com/olekukonko/tablewriter"
)

const (
	helpMessage = `
	h: help
	j: down cursor
	k: up cursor
	u: toggle the unit of bps or pps [][k][m]
	d: toggle the display of Down I/F
	p: toggle the display of bps or pps
	/: narrow down with regex
	   Targets of narrowing down are Description and I/F
	q: quit
	`
)

type Unit int

const (
	Bps Unit = iota
	Kbps
	Mbps
	Pps
	Kpps
	Mpps
)

func (u Unit) String() string {
	switch u {
	case Bps:
		return "bps"
	case Kbps:
		return "kbps"
	case Mbps:
		return "mbps"
	case Pps:
		return "pps"
	case Kpps:
		return "kpps"
	case Mpps:
		return "mpps"
	}
	return ""
}

type UnitCalc func(int64) int64

type MainWidget struct {
	Name          string
	Hosts         []*Host
	displayDownIF bool
	displaybps    bool
	unit          Unit
	unitCalc      UnitCalc
	log           *Logger
	*NarrowWidget
}

type NarrowWidget struct {
	Name   string
	regexp *regexp.Regexp
	log    *Logger
}

type Editor struct{}

func NewMainWidget(name string, hosts []*Host, nw *NarrowWidget, l *Logger) *MainWidget {
	m := &MainWidget{
		Name:          name,
		Hosts:         hosts,
		displayDownIF: true,
		displaybps:    true,
		NarrowWidget:  nw,
		log:           l,
	}
	if err := m.setUnit(Bps); err != nil {
		return nil
	}
	return m
}

func (m *MainWidget) setUnit(unit Unit) error {
	m.unit = unit
	switch m.unit {
	case Bps:
		m.unitCalc = func(x int64) int64 { return x * 8 }
	case Kbps:
		m.unitCalc = func(x int64) int64 { return x * 8 / 1024 }
	case Mbps:
		m.unitCalc = func(x int64) int64 { return x * 8 / 1024 / 1024 }
	case Pps:
		m.unitCalc = func(x int64) int64 { return x }
	case Kpps:
		m.unitCalc = func(x int64) int64 { return x / 1000 }
	case Mpps:
		m.unitCalc = func(x int64) int64 { return x / 1000 / 1000 }
	default:
		return fmt.Errorf("Unspecified value %v", m.unit)
	}
	return nil
}

func (m *MainWidget) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	v, err := g.SetView(m.Name, 0, 0, maxX-2, maxY-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
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
		"Stat",
		fmt.Sprintf("IN[%v]", m.unit.String()),
		fmt.Sprintf("OUT[%v]", m.unit.String()),
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

	for _, k := range keys {
		// Don't display Down I/F
		if !m.displayDownIF && h.IFs[k].OperStatus == "Down" {
			continue
		}
		s := fmt.Sprintf("%v %v", h.IFs[k].Desc, h.IFs[k].Alias)
		// toggle display bps or pps
		in := h.IFs[k].InOctets.Rate
		out := h.IFs[k].OutOctets.Rate
		if !m.displaybps {
			in = h.IFs[k].InUcastPkts.Rate
			out = h.IFs[k].OutUcastPkts.Rate
		}
		data := []string{
			h.Name,
			h.IFs[k].Desc,
			h.IFs[k].OperStatus,
			humanize.Comma(m.unitCalc(in)),
			humanize.Comma(m.unitCalc(out)),
			humanize.Comma(h.IFs[k].InError.Diff),
			humanize.Comma(h.IFs[k].OutError.Diff),
			humanize.Comma(h.IFs[k].InDiscards.Diff),
			humanize.Comma(h.IFs[k].OutDiscards.Diff),
			h.IFs[k].Alias,
		}
		if m.NarrowWidget.regexp.MatchString(s) {
			*narrowed = append(*narrowed, data)
		} else {
			*other = append(*other, data)
		}
	}
}

func NewNarrowWidget(name string, expr string, l *Logger) *NarrowWidget {
	n := &NarrowWidget{
		Name: name,
		log:  l,
	}
	if err := n.setRegexp(expr); err != nil {
		n.log.Warn().Msgf("failed setRegexp %v", err)
		return nil
	}
	return n
}

func (n *NarrowWidget) setRegexp(expr string) error {
	//Nothing is given, Nothing matches.
	if expr == "" {
		expr = "$^"
	}
	r, err := regexp.Compile(expr)
	if err != nil {
		return err
	}
	n.regexp = r
	return nil
}

func (w *NarrowWidget) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	v, err := g.SetView(w.Name, 0, maxY-2, maxX-2, maxY)
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
