package chinadns

import (
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"time"
)

type Options struct {
	Servers          string        `short:"s" required:"true" description:"Upstream dns servers, separated by ','"`
	DNSClientTimeout time.Duration `short:"t" default:"5s" description:"DNS query timeout in seconds"`
	ListenPort       string        `short:"p" default:"53"`
	ListenAddr       string        `short:"l" default:""`
}

func NewOptions() *Options {
	opt := new(Options)
	_, err := flags.Parse(opt)
	if err != nil {
		log.Fatal("Parse error:", err)
	}
	return opt
}
