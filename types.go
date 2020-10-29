package chinadns

import (
	"github.com/miekg/dns"
	"strconv"
)

type queryResult struct {
	server *ServerConfig
	ans    *dns.Msg
}

type dnsAnswer struct {
	ans *dns.Msg
	err error
}

type ServerConfig struct {
	IP        string
	Port      int
	IsCN      bool
	dnsClient *dns.Client
}

func (s *ServerConfig) String() string {
	return s.IP + ":" + strconv.Itoa(s.Port)
}
