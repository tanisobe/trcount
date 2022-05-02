package snmp

import (
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

const (
	ifDescr          string = ".3.6.1.2.1.2.2.1.2."
	ifAlias          string = ".3.6.1.2.1.31.1.1.1.18."
	ifSpeed          string = ".3.6.1.2.1.2.2.1.5."
	ifAdminStatus    string = ".3.6.1.2.1.2.2.1.7"
	ifOperStatus     string = ".3.6.1.2.1.2.2.1.8"
	ifHCInOctets     string = ".3.6.1.2.1.31.1.1.1.6."
	ifHCOutOctets    string = ".3.6.1.2.1.31.1.1.1.10."
	ifHCInUcastPkts  string = ".3.6.1.2.1.31.1.1.1.7."
	ifHCOutUcastPkts string = ".3.6.1.2.1.31.1.1.1.11."
	ifInDiscards     string = ".3.6.1.2.1.2.2.1.13."
	ifOutDiscards    string = ".3.6.1.2.1.2.2.1.19."
	ifInErrors       string = ".3.6.1.2.1.2.2.1.14."
	ifOutErrors      string = ".3.6.1.2.1.2.2.1.20."
)

type Host struct {
	Name    string
	IPAddr  *net.IPAddr
	IFs     map[int]*IF
	IFNum   int
	cstring *string
	dlog    *log.Logger
	params  *gosnmp.GoSNMP
}

type IF struct {
	Name         string
	Index        int
	Speed        int64
	Desc         string
	Alias        string
	AdminStatus  string
	OperStatus   string
	InOctets     *Counter
	OutOctets    *Counter
	InUcastPkts  *Counter
	OutUcastPkts *Counter
	InDiscards   *Counter
	OutDiscards  *Counter
	InError      *Counter
	OutError     *Counter
}

type Counter struct {
	Last       int64
	Before     int64
	LastTime   time.Time
	BeforeTime time.Time
	Diff       int64
	Rate       int64
}

type Option func(*Host) error

func DLogger(dl *log.Logger) Option {
	return func(h *Host) error {
		h.dlog = dl
		return nil
	}
}
func Cstring(c *string) Option {
	return func(h *Host) error {
		h.cstring = c
		return nil
	}
}
func NewHost(hostname string, options ...Option) *Host {
	h := new(Host)
	ipaddr, err := net.ResolveIPAddr("ip", hostname)
	if err != nil {
		log.Fatalln(err)
	}
	h.IPAddr = ipaddr
	h.Name = hostname
	h.IFs = make(map[int]*IF)
	h.dlog = log.New(ioutil.Discard, "", log.LstdFlags|log.LUTC)
	for _, option := range options {
		option(h)
	}
	return h
}

func NewIF(index int) *IF {
	i := new(IF)
	i.Index = index
	i.AdminStatus = ""
	i.OperStatus = ""
	i.InOctets = new(Counter)
	i.OutOctets = new(Counter)
	i.InUcastPkts = new(Counter)
	i.OutUcastPkts = new(Counter)
	i.InDiscards = new(Counter)
	i.OutDiscards = new(Counter)
	i.InError = new(Counter)
	i.OutError = new(Counter)
	return i
}

func (h *Host) InitHost() error {
	h.dlog.Println("Connect")
	h.params = &gosnmp.GoSNMP{
		Target:    h.IPAddr.IP.String(),
		Port:      161,
		Version:   gosnmp.Version2c,
		Community: *h.cstring,
		Logger:    gosnmp.NewLogger(h.dlog),
		Timeout:   time.Duration(3) * time.Second,
	}
	if err := h.params.Connect(); err != nil {
		h.dlog.Fatalf("Connect() err: %v", err)
		return err
	}
	defer h.params.Conn.Close()

	//GET ALL Interface Index
	h.dlog.Printf("Get ALL Interface Index")
	oid := ".1.3.6.1.2.1.2.2.1.1"
	if err := h.params.BulkWalk(oid, h.StoreValue); err != nil {
		h.dlog.Printf("Get() err: %v", err)
		return err
	}
	//GET ALL Interface Value
	h.dlog.Printf("Get ALL Interface Value")
	oid = ".1.3.6.1.2.1.2"
	if err := h.params.BulkWalk(oid, h.UpdateIFValue); err != nil {
		h.dlog.Printf("Get() err: %v", err)
		return err
	}

	//GET IF MIB
	h.dlog.Printf("Get ALL Interface Value")
	oid = ".1.3.6.1.2.1.31.1.1.1"
	if err := h.params.BulkWalk(oid, h.UpdateIFValue); err != nil {
		h.dlog.Printf("Get() err: %v", err)
		return err
	}
	return nil
}

func (h *Host) Update() {
	h.dlog.Printf("Update %v", h.Name)
	if err := h.params.Connect(); err != nil {
		h.dlog.Printf("Connect() err: %v", err)
	}
	defer h.params.Conn.Close()

	//GET ALL Interface Value
	oid := ".1.3.6.1.2.1.2.2"
	if err := h.params.BulkWalk(oid, h.UpdateIFValue); err != nil {
		h.dlog.Printf("Get() err: %v", err)
	}
	oid = ".1.3.6.1.2.1.31.1.1.1"
	if err := h.params.BulkWalk(oid, h.UpdateIFValue); err != nil {
		h.dlog.Printf("Get() err: %v", err)
	}
}

func (h *Host) StoreValue(pdu gosnmp.SnmpPDU) error {

	index := int(gosnmp.ToBigInt(pdu.Value).Int64())
	h.IFs[index] = NewIF(index)

	return nil
}

// Retrive data and data
func (h *Host) UpdateIFValue(pdu gosnmp.SnmpPDU) error {
	h.dlog.Println(pdu)
	s := strings.Split(pdu.Name, ".")
	index, _ := strconv.Atoi(s[len(s)-1])
	t := time.Now()
	h.dlog.Printf("UpdateIFValue %v", t)

	switch {
	case strings.Contains(pdu.Name, ifDescr):
		h.IFs[index].Desc = string(pdu.Value.([]byte))
	case strings.Contains(pdu.Name, ifAlias):
		h.IFs[index].Alias = string(pdu.Value.([]byte))
	case strings.Contains(pdu.Name, ifSpeed):
		h.IFs[index].Speed = gosnmp.ToBigInt(pdu.Value).Int64()
	case strings.Contains(pdu.Name, ifAdminStatus):
		switch gosnmp.ToBigInt(pdu.Value).Int64() {
		case 1:
			h.IFs[index].AdminStatus = "UP"
		case 2:
			h.IFs[index].AdminStatus = "Down"
		}
	case strings.Contains(pdu.Name, ifOperStatus):
		switch gosnmp.ToBigInt(pdu.Value).Int64() {
		case 1:
			h.IFs[index].OperStatus = "UP"
		case 2:
			h.IFs[index].OperStatus = "Down"
		}
	case strings.Contains(pdu.Name, ifHCInOctets):
		h.IFs[index].InOctets.Update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifHCOutOctets):
		h.IFs[index].OutOctets.Update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifHCInUcastPkts):
		h.IFs[index].InUcastPkts.Update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifHCOutUcastPkts):
		h.IFs[index].OutUcastPkts.Update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifInDiscards):
		h.IFs[index].InDiscards.Update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifOutDiscards):
		h.IFs[index].OutDiscards.Update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifInDiscards):
		h.IFs[index].InDiscards.Update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifOutErrors):
		h.IFs[index].OutError.Update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifInErrors):
		h.IFs[index].InError.Update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	}
	return nil
}

func (c *Counter) Update(v int64, t time.Time) {
	c.BeforeTime = c.LastTime
	c.Before = c.Last
	c.LastTime = t
	c.Last = v
	c.Diff = c.Last - c.Before
	d := c.LastTime.Sub(c.BeforeTime)
	c.Rate = c.Diff / int64(d.Seconds())
}
