package dns

import (
	"encoding/json"
	"fmt"
	"github.com/miekg/dns"
	"github.com/ruijzhan/chinadns-go/cidr"
	"github.com/ruijzhan/chinadns-go/options"
	"github.com/yl2chen/cidranger"
	"io/ioutil"
	"log"
	"os"
)

var (
	dnsClient *dns.Client
	chnRanger cidranger.Ranger
	opt       *options.Options
	mType     = map[uint16]string{
		dns.TypeA:     "A",
		dns.TypeCNAME: "CNAME",
		dns.TypePTR:   "PTR",
		dns.TypeAAAA:  "AAAA",
		dns.TypeSRV:   "SRV"}
)

func init() {
	dnsClient = new(dns.Client)
	opt = &options.Opt
	dnsClient.Timeout = opt.DNSClientTimeout

	f, err := os.Open(opt.CHNRoutePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	cidrs := new(cidr.CIDRs)
	json.Unmarshal(b, cidrs)
	chnRanger = cidr.GetCHNRanger(cidrs)
}

func isChineseARecord(msg *dns.Msg) (bool, error) {
	for _, rr := range msg.Answer {
		if rr, ok := rr.(*dns.A); ok {
			cn, err := chnRanger.Contains(rr.A)
			if err != nil {
				log.Println(err)
			}
			return cn, nil
		}
	}
	if msg.Question[0].Qtype == dns.TypeAAAA {
		return false, nil
	}
	return false, fmt.Errorf("query to %s returns %d ANS, %d AUTH, %d EXT",
		msg.Question[0].Name, len(msg.Answer), len(msg.Ns), len(msg.Extra))
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	if r.Opcode == dns.OpcodeQuery {
		m := doubleDNSQuery(r)
		w.WriteMsg(m)
		return
	}
	log.Printf("Unsupported Opcode: %d", r.Opcode)
	w.WriteMsg(&dns.Msg{})
}

func makeQuery(request *dns.Msg, dnsServer string, resultCh chan<- *dns.Msg) {
	m := &dns.Msg{}
	m.SetReply(request)
	m.Compress = true
	for _, q := range m.Question {
		req := &dns.Msg{}
		req.SetQuestion(q.Name, q.Qtype)
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

	go makeQuery(request, opt.ChinaDNSServer, chCN)
	go makeQuery(request, opt.NoneChinaDNSServer, chI8l)

	cnMsg := <-chCN
	if cnIP, err := isChineseARecord(cnMsg); err != nil {
		log.Println(err)
		return cnMsg
	} else {
		if cnIP {
			log.Printf("Returned result from %s: %s\n", opt.ChinaDNSServer, request.Question[0].Name)
			return cnMsg
		}
	}

	log.Printf("Returned result from %s: %s\n", opt.NoneChinaDNSServer, request.Question[0].Name)
	return <-chI8l
}

func RunDNSServer() error {
	dns.HandleFunc(".", handleDnsRequest)
	server := &dns.Server{Addr: options.Opt.ListenAddr + ":" + options.Opt.ListenPort, Net: "udp"}
	defer server.Shutdown()

	log.Printf("Starting at %s\n", options.Opt.ListenPort)
	return server.ListenAndServe()
}
