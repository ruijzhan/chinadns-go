package dns

import (
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"time"
)

func requestHandler(w dns.ResponseWriter, r *dns.Msg) {
	if r.Opcode == dns.OpcodeQuery {
		m := multiQuery(r, opts.DNSServers, w.RemoteAddr(), 10*time.Second)
		w.WriteMsg(m)
	} else {
		log.Warningf("Unsupported Opcode: %d", r.Opcode)
		w.WriteMsg(&dns.Msg{})
	}
}

func RunDNSServer() error {
	dns.HandleFunc(".", requestHandler)
	server := &dns.Server{Addr: opts.ListenAddr + ":" + opts.ListenPort, Net: "udp"}
	defer server.Shutdown()

	log.Infof("Starting at %s\n", opts.ListenPort)
	return server.ListenAndServe()
}
