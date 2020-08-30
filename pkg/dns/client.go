package dns

import (
	"github.com/miekg/dns"
	"log"
)

var	dnsClient *dns.Client

func init()  {
	dnsClient = new(dns.Client)
}

func makeQuery(request *dns.Msg, dnsServer string, resultCh chan<- *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(request)
	m.Compress = true
	for _, q := range m.Question {
		req := &dns.Msg{}
		req.SetQuestion(q.Name, q.Qtype)
		dnsClient.Timeout = opts.DNSClientTimeout
		ans, _, err := dnsClient.Exchange(req, dnsServer)
		if err != nil {
			log.Printf("Resolving %s error: %v\n", q.Name, err)
		} else {
			for _, rr := range ans.Answer {
				m.Answer = append(m.Answer, rr)
			}
			for _, rr := range ans.Ns {
				m.Ns = append(m.Ns, rr)
			}
			for _, rr := range ans.Extra {
				m.Extra = append(m.Extra, rr)
			}
		}
	}
	resultCh <- m
}

func doubleDNSQuery(request *dns.Msg) *dns.Msg {
	chCN := make(chan *dns.Msg, 1)
	chI8l := make(chan *dns.Msg, 1)

	go makeQuery(request, opts.ChinaDNSServer, chCN)
	go makeQuery(request, opts.NoneChinaDNSServer, chI8l)

	cnMsg := <-chCN
	if cnIP, err := isChineseARecord(cnMsg); err != nil {
		log.Println(err)
		return cnMsg
	} else {
		if cnIP {
			log.Printf("Returned result from %s: %s\n", opts.ChinaDNSServer, request.Question[0].Name)
			return cnMsg
		}
	}

	log.Printf("Returned result from %s: %s\n", opts.NoneChinaDNSServer, request.Question[0].Name)
	return <-chI8l
}