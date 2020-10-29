package chinadns

import (
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/uniplaces/carbon"
)

var (
	dnsServers []*ServerConfig
)

func requestHandler(w dns.ResponseWriter, r *dns.Msg) {
	if r.Opcode == dns.OpcodeQuery {
		start := carbon.Now().Time
		m, err := queryServers(r, dnsServers)
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
				r.Question[0].Name[:len(r.Question[0].Name)-1],
				msTaken,
			)
		}

	} else {
		log.Warningf("Unsupported Opcode: %d", r.Opcode)
		w.WriteMsg(&dns.Msg{})
	}
}

func RunDNSServer(opts *Options) error {
	dnsServers = opts.DNSServers
	dns.HandleFunc(".", requestHandler)
	server := &dns.Server{Addr: opts.ListenAddr + ":" + opts.ListenPort, Net: "udp"}
	defer server.Shutdown()

	log.Infof("Starting at %s\n", opts.ListenPort)
	return server.ListenAndServe()
}
