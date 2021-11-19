package snmp_test

import (
	"testing"
	"time"

	"github.com/tanisobe/trcount/snmp"
)

func Test_NewHost(t *testing.T) {
	h := snmp.NewHost("127.0.0.1")

	h.InitHost()

	h.Update()
}

func Test_MultiUpdateHost(t *testing.T) {
	a := snmp.NewHost("127.0.0.1")
	b := snmp.NewHost("localhost")

	a.InitHost()
	b.InitHost()

	go func() {
		a.Update()
	}()

	go func() {
		b.Update()
	}()

	time.Sleep(30 * time.Second)
}
