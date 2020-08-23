package options

import (
	"github.com/jessevdk/go-flags"
	"log"
	"time"
)

var (
	Opt Options
)

func init() {
	_, err := flags.Parse(&Opt)
	if err != nil {
		log.Fatal("Parse error:", err)
	}
}

type Options struct {
	ChinaDNSServer     string        `short:"c" required:"true"`
	NoneChinaDNSServer string        `short:"n" required:"true"`
	DNSClientTimeout   time.Duration `short:"t" default:"5s" description:"DNS query timeout in seconds"`
	CHNRoutePath       string        `short:"f" default:"chnroute.json" description:"Path to chnroute file"`
	ListenPort         string        `short:"p" default:"53"`
	ListenAddr         string        `short:"l" default:""`
}
