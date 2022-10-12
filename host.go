package trmon

import (
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

	ifIndex  string = ".1.3.6.1.2.1.2.2.1.1"
	ifEntry  string = ".1.3.6.1.2.1.2.2.1"
	ifXEntry string = ".1.3.6.1.2.1.31.1.1.1"
)

type Host struct {
	Name   string
	IFs    map[int]*IF
	params *gosnmp.GoSNMP
	log    *Logger
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
	name       string
	Last       int64
	Before     int64
	LastTime   time.Time
	BeforeTime time.Time
	Diff       int64
	Rate       int64
	log        *Logger
}

func newCounter(name string, log *Logger) *Counter {
	return &Counter{
		name: name,
		log:  log,
	}
}

func newIF(index int, l *Logger) *IF {
	i := new(IF)
	i.Index = index
	i.AdminStatus = ""
	i.OperStatus = ""
	i.InOctets = newCounter("InOctets", l)
	i.OutOctets = newCounter("OutOctets", l)
	i.InUcastPkts = newCounter("InUcastPkts", l)
	i.OutUcastPkts = newCounter("OutUcastPkts", l)
	i.InDiscards = newCounter("InDiscards", l)
	i.OutDiscards = newCounter("OutDiscards", l)
	i.InError = newCounter("InError", l)
	i.OutError = newCounter("OutError", l)
	return i
}

func (h *Host) newIFs(pdu gosnmp.SnmpPDU) error {

	index := int(gosnmp.ToBigInt(pdu.Value).Int64())
	h.IFs[index] = newIF(index, h.log)

	return nil
}

func NewHost(hostname string, community string, l *Logger) (*Host, error) {
	h := &Host{
		Name: hostname,
		IFs:  make(map[int]*IF),
		params: &gosnmp.GoSNMP{
			Target:    hostname,
			Port:      161,
			Version:   gosnmp.Version2c,
			Community: community,
			Timeout:   time.Duration(3) * time.Second,
		},
		log: l,
	}

	if err := h.params.Connect(); err != nil {
		h.log.Debug().Msgf("Connect() err: %v", err)
		return nil, err
	}
	defer h.params.Conn.Close()

	//GET ALL Interface Index
	h.log.Debug().Msg("Get ALL Interface Index")
	if err := h.params.BulkWalk(ifIndex, h.newIFs); err != nil {
		h.log.Debug().Msgf("Failed to new IFs: %v", err)
		return nil, err
	}
	return h, nil
}

func (h *Host) Update() {
	h.log.Debug().Msgf("Update IFs %v", h.Name)
	if err := h.params.Connect(); err != nil {
		h.log.Debug().Msgf("Connect() err: %v", err)
	}
	defer h.params.Conn.Close()

	//GET ALL Interface Value
	if err := h.params.BulkWalk(ifEntry, h.updateIFValue); err != nil {
		h.log.Debug().Msgf("Failed to Update ifEntry: %v", err)
	}
	if err := h.params.BulkWalk(ifXEntry, h.updateIFValue); err != nil {
		h.log.Debug().Msgf("Failed to Update ifXEntry: %v", err)
	}
}

// Classify retrived snmp PDU and Set new value to IF array
func (h *Host) updateIFValue(pdu gosnmp.SnmpPDU) error {
	h.log.Debug().Msgf("pdu %v", pdu)
	s := strings.Split(pdu.Name, ".")
	index, _ := strconv.Atoi(s[len(s)-1])
	t := time.Now()

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
		h.IFs[index].InOctets.update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifHCOutOctets):
		h.IFs[index].OutOctets.update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifHCInUcastPkts):
		h.IFs[index].InUcastPkts.update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifHCOutUcastPkts):
		h.IFs[index].OutUcastPkts.update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifInDiscards):
		h.IFs[index].InDiscards.update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifOutDiscards):
		h.IFs[index].OutDiscards.update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifInDiscards):
		h.IFs[index].InDiscards.update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifOutErrors):
		h.IFs[index].OutError.update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	case strings.Contains(pdu.Name, ifInErrors):
		h.IFs[index].InError.update(gosnmp.ToBigInt(pdu.Value).Int64(), t)
	}
	return nil
}

func (c *Counter) update(v int64, t time.Time) {
	c.BeforeTime = c.LastTime
	c.Before = c.Last
	c.LastTime = t
	c.Last = v
	//When the counter goes around
	if d := c.Last - c.Before; d < 0 {
		c.log.Debug().Msgf("the counter %v goes around", c.name)
		c.Diff = 0
	} else {
		c.Diff = d
	}
	//When the time difference is less than one second
	d := int64((c.LastTime.Sub(c.BeforeTime)).Seconds())
	if d == 0 {
		c.log.Debug().Msgf("zero devide %v", c.name)
		c.Rate = 0
	} else {
		c.Rate = c.Diff / d
	}
}
