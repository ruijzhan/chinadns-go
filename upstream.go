package chinadns

import (
	"fmt"
	"github.com/miekg/dns"
	cidr "github.com/ruijzhan/country-cidr"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type remoteDNS struct {
	IP        string
	Port      int
	IsCN      bool
	dnsClient *dns.Client
}

func parseServers(s string) ([]*remoteDNS, error) {
	servers := make([]*remoteDNS, 0)
	for _, serverStr := range strings.Split(s, ",") {
		ts := strings.Split(serverStr, ":")
		isCN := cidr.Country("CN").ContainsIPstr(ts[0])
		server := &remoteDNS{
			IP:        ts[0],
			Port:      53,
			IsCN:      isCN,
			dnsClient: new(dns.Client),
		}
		if len(ts) == 1 {
		} else if len(ts) == 2 {
			port, err := strconv.Atoi(ts[1])
			if err != nil {
				return servers, fmt.Errorf("invalid port number: %v", ts[1])
			}
			server.Port = port
		} else {
			return servers, fmt.Errorf("invalid server address: %s", serverStr)
		}
		servers = append(servers, server)
	}
	return servers, nil
}

func (svr remoteDNS) String() string {
	return svr.IP + ":" + strconv.Itoa(svr.Port)
}

// answer resolves a single dns question
func (svr remoteDNS) answer(question dns.Question) (*DNSResult, error) {
	question.Name = rmHttp(question.Name)
	req := &dns.Msg{}
	req = req.SetQuestion(question.Name, question.Qtype)
	req.Compress = true
	ans, _, err := svr.dnsClient.Exchange(req, svr.String())
	if err != nil {
		return nil, err
	}
	return &DNSResult{
		server: svr,
		answer: ans,
	}, nil
}

// resolve resolves a dns request in an async way
func (svr remoteDNS) resolve(req *dns.Msg) <-chan *DNSResult {
	chResults := make(chan *DNSResult, 1)
	go func() {
		defer close(chResults)
		m := &dns.Msg{}
		m.SetReply(req)
		if len(m.Question) > 0 {
			q := m.Question[0]
			res, err := svr.answer(q)
			if err != nil {
				log.Warn(err)
				chResults <- &DNSResult{err: err}
			} else {
				m.Answer = append(m.Answer, res.answer.Answer...)
				m.Ns = append(m.Ns, res.answer.Ns...)
				m.Extra = append(m.Extra, res.answer.Extra...)
				if len(m.Answer) > 0 {
					chResults <- &DNSResult{svr, m, nil}
				} else {
					//log.Warn("Got no answer")
					chResults <- &DNSResult{err: fmt.Errorf("no answer error: %v", q)}
				}
			}
		}
	}()

	return chResults
}
