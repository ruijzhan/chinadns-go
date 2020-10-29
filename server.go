package chinadns

import (
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/uniplaces/carbon"
	"time"
)

type ChinaDNS struct {
	upstreamServs []*ServerConfig
	servTimeout   time.Duration
}

func NewChinaDNS(opt *Options) *ChinaDNS {
	servers := parseServers(opt.Servers)
	for _, s := range servers {
		s.dnsClient = new(dns.Client)
		s.dnsClient.Timeout = opt.DNSClientTimeout
	}

	return &ChinaDNS{upstreamServs: servers,
		servTimeout: opt.DNSClientTimeout,
	}
}

func (c *ChinaDNS) requestHandler(w dns.ResponseWriter, question *dns.Msg) {
	if question.Opcode == dns.OpcodeQuery {
		start := carbon.Now().Time
		m, err := c.resolve(question)
		msTaken := carbon.Now().Sub(start).Milliseconds()
		// Reply to client anyway
		w.WriteMsg(m)

		if err != nil {
			log.Warning(err)
		} else {
			isCN, _ := isChineseARecord(m)
			log.Infof("[%s] -> [ChinaIP: %v] [%s] %dms",
				w.RemoteAddr().String(),
				isCN,
				question.Question[0].Name[:len(question.Question[0].Name)-1],
				msTaken,
			)
		}

	} else {
		log.Warningf("Unsupported Opcode: %d", question.Opcode)
		w.WriteMsg(&dns.Msg{})
	}
}

func RunDNSServer(opts *Options) error {
	chinaDNS := NewChinaDNS(opts)
	dns.HandleFunc(".", chinaDNS.requestHandler)
	server := &dns.Server{Addr: opts.ListenAddr + ":" + opts.ListenPort, Net: "udp"}
	defer server.Shutdown()

	log.Infof("Starting at %s\n", opts.ListenPort)
	return server.ListenAndServe()
}
