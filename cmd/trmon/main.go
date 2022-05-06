package main

import (
	"flag"
	"fmt"
	"io"
	"runtime/debug"
	"time"

	"log"
	"os"

	"github.com/tanisobe/trmon"
)

var (
	version  string = "unkown"
	revision string = "unkown"
)

func main() {
	e := flag.String("e", "", `narrow down to IFs that match with regular expressions.
	Matching target is IF name and IF Description`)
	d := flag.Bool("debug", false, "start with debug mode. deubg mode dump trace log")
	c := flag.String("c", "public", "snmp community string.")
	i := flag.Int("i", 5, "SNMP polling interval [sec]. minimum 5")
	l := flag.Int("l", 7200, "trmon continuous operation time [sec].")
	v := flag.Bool("v", false, "show app version")
	flag.Parse()

	if *v {
		info, ok := debug.ReadBuildInfo()
		if ok {
			version = info.Main.Version
		}
		fmt.Printf("trmon: traffic monitor\nversion: %s\nrevision:%s\n", version, revision)
		os.Exit(0)
	}

	if *i < 5 {
		log.Println("Too short interval, The minimum SNMP polling interval is 5 seconds")
		os.Exit(1)
	}

	if len(flag.Args()) < 1 {
		log.Println("Must specify at least one host")
		os.Exit(1)
	}

	var f io.Writer
	if *d {
		file, err := os.Create(fmt.Sprintf("trmon%v.log", time.Now().Unix()))
		defer file.Close()
		f = file

		if err != nil {
			log.Printf("Failed to create log file: %v", err)
			os.Exit(1)
		}
	} else {
		f = os.Stderr
	}

	config := &trmon.Config{
		Interval:  *i,
		Lifespan:  *l,
		Community: *c,
		Expr:      *e,
		IsDebug:   *d,
		Output:    f,
	}

	app := new(trmon.App)
	app.Run(flag.Args(), config)
}
