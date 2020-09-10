package dns

import (
	"github.com/miekg/dns"
	"log"
	"time"
)

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	if r.Opcode == dns.OpcodeQuery {
		m := multiQuery(r, opts.DNSServers, w.RemoteAddr(), 10*time.Second)
		w.WriteMsg(m)
	} else {
		log.Printf("Unsupported Opcode: %d", r.Opcode)
		w.WriteMsg(&dns.Msg{})
	}
}

func RunDNSServer() error {
	dns.HandleFunc(".", handleDnsRequest)
	server := &dns.Server{Addr: opts.ListenAddr + ":" + opts.ListenPort, Net: "udp"}
	defer server.Shutdown()

	log.Printf("Starting at %s\n", opts.ListenPort)
	return server.ListenAndServe()
}
