package dns

import (
	"fmt"
	"github.com/miekg/dns"
	"github.com/ruijzhan/chinadns-go/pkg/options"
	"log"
)

var	dnsClient *dns.Client

func init()  {
	dnsClient = new(dns.Client)
}

type queryResult struct {
	serverConfig *options.ServerConfig
	dnsResult    *dns.Msg
}

func singleQuery(request *dns.Msg, dnsServer *options.ServerConfig, resultCh chan<- *queryResult) {
	m := &dns.Msg{}
	m.SetReply(request)
	m.Compress = true
	for _, q := range m.Question {
		req := &dns.Msg{}
		req.SetQuestion(q.Name, q.Qtype)
		dnsClient.Timeout = opts.DNSClientTimeout
		ans, _, err := dnsClient.Exchange(req, dnsServer.String())
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
	resultCh <- &queryResult{dnsServer,m}
}

func waitResult(ch <-chan *queryResult) (*queryResult, error) {
	var saved *queryResult
	var confirmed bool

	for result := range ch {
		isChinaDNS, err := isChineseIP(result.serverConfig.IP)
		if err != nil {
			log.Println(err)
			continue
		}
		isChinaResult, err := isChineseARecord(result.dnsResult)
		if err != nil {
			log.Println(err)
			continue
		}

		switch {
		case isChinaDNS && isChinaResult:
			return result,nil //中国解析IP 直接返回
		case isChinaDNS && !isChinaResult:
			confirmed = true //确认不是中国解析结果
			if saved != nil {
				return saved,nil //如果外国服务器结果已经保存，直接返回
			}
		case !isChinaDNS && isChinaResult:
			return result,nil //中国解析结果，直接返回
		case !isChinaDNS && !isChinaResult:
			if confirmed { //如果中国服务器确定是外国IP，直接返回结果
				return result,nil
			} else {
				if saved != nil {
					saved = result //否则缓存起来
				}
			}
		}
	}
	return nil, fmt.Errorf("failed to resolve domain on all servers")
}

func multiQuery(request *dns.Msg, servers []*options.ServerConfig) *dns.Msg {
	chResults := make(chan *queryResult, len(servers))

	for _, server := range servers {
		go singleQuery(request, server, chResults)
	}

	result, err := waitResult(chResults)
	if err != nil {
		log.Println(err)
		return &dns.Msg{}
	}
	log.Printf("Resulted result from %s: %s\n", result.serverConfig.IP, result.dnsResult.Question[0].Name)
	return result.dnsResult
}