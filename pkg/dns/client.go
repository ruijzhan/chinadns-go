package dns

import (
	"context"
	"fmt"
	"github.com/miekg/dns"
	"github.com/ruijzhan/chinadns-go/pkg/options"
	"github.com/uniplaces/carbon"
	"log"
	"net"
	"sync"
	"time"
)

var dnsClient *dns.Client

func init() {
	dnsClient = new(dns.Client)
}

type queryResult struct {
	server   *options.ServerConfig
	resolved *dns.Msg
	isCN     bool
	err      error
}

type exchangeResult struct {
	ans *dns.Msg
	rtt time.Duration
	err error
}

func asyncExchange(cli *dns.Client, req *dns.Msg, server string, ctx context.Context) <-chan *exchangeResult {
	ch := make(chan *exchangeResult)

	go func() {
		ans, rtt, err := cli.Exchange(req, server)
		select {
		case ch <- &exchangeResult{
			ans: ans,
			rtt: rtt,
			err: err,
		}:
		case <-ctx.Done():
			return
		}

	}()
	return ch
}

func singleQuery(request *dns.Msg, dnsServer *options.ServerConfig, resultCh chan<- *queryResult, ctx context.Context) {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	m := &dns.Msg{}
	m.SetReply(request)
	for _, q := range m.Question {
		wg.Add(1)
		go func(q dns.Question) {
			defer wg.Done()
			q.Name = rmHttp(q.Name)
			req := &dns.Msg{}
			req = req.SetQuestion(q.Name, q.Qtype)
			req.Compress = true
			dnsClient.Timeout = opts.DNSClientTimeout
			ch := asyncExchange(dnsClient, req, dnsServer.String(), ctx)
			select {
			case res := <-ch:
				if res.err != nil {
					log.Printf("Resolving %s error: %v\n", q.Name, res.err)
				} else {
					mu.Lock()
					m.Answer = append(m.Answer, res.ans.Answer...)
					m.Ns = append(m.Ns, res.ans.Ns...)
					m.Extra = append(m.Extra, res.ans.Extra...)
					mu.Unlock()
				}
			case <-ctx.Done():
				return
			}
		}(q)
	}
	wg.Wait()
	if len(m.Answer) == 0 { //canceled by context
		return
	}
	if isCN, err := isChineseARecord(m); err == nil {
		resultCh <- &queryResult{dnsServer, m, isCN, nil}
	}
}

func filter(ch <-chan *queryResult) (*queryResult, error) {
	var cached *queryResult
	var confirmed bool
	var servers []string

	for result := range ch {
		if result.err != nil {
			//log.Println(result.err)
			continue
		}
		servers = append(servers, result.server.IP)
		switch {
		case result.server.IsCN && result.isCN:
			return result, nil //中国解析IP 直接返回
		case result.server.IsCN && !result.isCN:
			if cached != nil {
				return cached, nil //如果外国服务器结果已经保存，直接返回
			}
			confirmed = true //确认不是中国解析结果
		case !result.server.IsCN && result.isCN:
			return result, nil //中国解析结果，直接返回
		case !result.server.IsCN && !result.isCN:
			if confirmed { //如果中国服务器确定是外国IP，直接返回结果
				return result, nil
			} else {
				if cached == nil {
					cached = result //否则缓存起来
				}
			}
		} //end of switch
	}
	return nil, fmt.Errorf("query timeout, servers responded: %v", servers)
}

func multiQuery(request *dns.Msg, servers []*options.ServerConfig, remoteAddr net.Addr, timeout time.Duration) *dns.Msg {
	start := carbon.Now().Time
	chResults := make(chan *queryResult, len(servers))
	go func() {
		<-time.After(timeout)
		close(chResults)
	}()

	ctx, cancel := context.WithCancel(context.Background())

	for _, server := range servers {
		go singleQuery(request, server, chResults, ctx)
	}

	result, err := filter(chResults)
	cancel()
	if err != nil {
		log.Printf("Resolving %s error: %s\n", request.Question[0].Name, err.Error())
		return &dns.Msg{}
	}
	go func() {
		msTaken := carbon.Now().Sub(start).Milliseconds()

		var country string
		if result.isCN {
			country = "CN"
		} else {
			country = "!CN"
		}

		log.Printf("[%s] -> [%s] [%s][%s] \t%dms",
			remoteAddr.String(),
			result.server.IP,
			country,
			result.resolved.Question[0].Name[:len(result.resolved.Question[0].Name)-1],
			msTaken,
		)
	}()

	return result.resolved
}
