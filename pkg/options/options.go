package options

import (
	"github.com/jessevdk/go-flags"
	"log"
	"time"
)

type Options struct {
	ChinaDNSServer     string        `short:"c" required:"true"`
	NoneChinaDNSServer string        `short:"n" required:"true"`
	DNSClientTimeout   time.Duration `short:"t" default:"5s" description:"DNS query timeout in seconds"`
	CHNRoutePath       string        `short:"f" default:"chnroute.json" description:"Path to chnroute file"`
	ListenPort         string        `short:"p" default:"53"`
	ListenAddr         string        `short:"l" default:""`
}

func NewOptions() *Options {
	opt := new(Options)
	_, err := flags.Parse(opt)
	if err != nil {
		log.Fatal("Parse error:", err)
	}
	return opt
}
