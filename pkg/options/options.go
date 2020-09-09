package options

import (
	"github.com/jessevdk/go-flags"
	"log"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	Servers          string        `short:"s" required:"true" description:"Upstream dns servers, separated by ','"`
	DNSClientTimeout time.Duration `short:"t" default:"5s" description:"DNS query timeout in seconds"`
	ListenPort       string        `short:"p" default:"53"`
	ListenAddr       string        `short:"l" default:""`
	DNSServers       []*ServerConfig
}

type ServerConfig struct {
	IP   string
	Port int
}

func (sc *ServerConfig) String() string {
	return sc.IP + ":" + strconv.Itoa(sc.Port)
}

func parseServers(s string) []*ServerConfig {
	servers := make([]*ServerConfig, 0)
	for _, serverStr := range strings.Split(s, ",") {
		ts := strings.Split(serverStr, ":")
		if len(ts) == 1 {
			servers = append(servers, &ServerConfig{IP: ts[0], Port: 53})
		} else if len(ts) == 2 {
			port, err := strconv.Atoi(ts[1])
			if err != nil {
				log.Fatalf("Invalid port number: %v", err)
			}
			servers = append(servers, &ServerConfig{IP: ts[0], Port: port})
		} else {
			log.Fatalf("Invalid server address: %s", s)
		}
	}
	return servers
}

func NewOptions() *Options {
	opt := new(Options)
	_, err := flags.Parse(opt)
	if err != nil {
		log.Fatal("Parse error:", err)
	}

	opt.DNSServers = parseServers(opt.Servers)
	return opt
}
