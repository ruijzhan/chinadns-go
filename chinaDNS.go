package chinadns

import (
	"context"
	"fmt"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/uniplaces/carbon"
	"time"
)

type queryResult struct {
	// server represents the upstream DNS server which
	// resolves the question
	server *remoteDNS
	answer *dns.Msg
	err    error
}

// filter validates dns resolving results read from chRes
// based on rules(see comments). It returns a valid result if there is any
// or an error.
func filter(chRes <-chan *queryResult) (*queryResult, error) {
	var cached *queryResult
	var confirmed bool
	var servers []string

	for res := range chRes {
		servers = append(servers, res.server.IP)
		ansIsCN, err := isChineseARecord(res.answer)
		// isChineseARecord may fail if the *dns.Msg contains no A records
		if err != nil {
			log.Debug(err)
			continue
		}

		switch {
		case res.server.IsCN && ansIsCN:
			return res, nil //中国解析IP 直接返回
		case res.server.IsCN && !ansIsCN:
			if cached != nil {
				return cached, nil //如果外国服务器结果已经保存，直接返回
			}
			confirmed = true //确认不是中国解析结果
		case !res.server.IsCN && ansIsCN:
			return res, nil //中国解析结果，直接返回
		case !res.server.IsCN && !ansIsCN:
			if confirmed { //如果中国服务器确定是外国IP，直接返回结果
				return res, nil
			} else {
				if cached == nil {
					cached = res //否则缓存起来
				}
			}
		}
	}
	return nil, fmt.Errorf("query timeout, servers responded: %v", servers)
}

type ChinaDNS struct {
	upstreamServs []*remoteDNS
	servTimeout   time.Duration
}

func NewChinaDNS(opt *Options) *ChinaDNS {
	servers := parseServers(opt.Servers)
	for _, s := range servers {
		s.dnsClient.Timeout = opt.DNSClientTimeout
	}

	return &ChinaDNS{upstreamServs: servers,
		servTimeout: opt.DNSClientTimeout,
	}
}

// resolve method sends dns question to multiple upstream dns servers
// filters a valid result and returns it
func (c *ChinaDNS) resolve(question *dns.Msg) (*dns.Msg, error) {
	chResults := make(chan *queryResult, len(c.upstreamServs))
	go func() {
		<-time.After(c.servTimeout)
		close(chResults)
	}()

	ctx, cancel := context.WithCancel(context.Background())

	for _, server := range c.upstreamServs {
		go server.resolve(question, chResults, ctx)
	}

	result, err := filter(chResults)
	cancel()

	if err != nil {
		return &dns.Msg{}, fmt.Errorf("Resolving %s error: %v\n", question.Question[0].Name, err)
	}
	return result.answer, nil
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
