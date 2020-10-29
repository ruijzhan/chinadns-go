package chinadns

import (
	"context"
	"fmt"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var dnsClient *dns.Client

func init() {
	dnsClient = new(dns.Client)
}

func (s *ServerConfig) query(req *dns.Msg, ctx context.Context) <-chan *dnsAnswer {
	ch := make(chan *dnsAnswer)

	go func() {
		ans, _, err := dnsClient.Exchange(req, s.String())
		select {
		case ch <- &dnsAnswer{
			ans: ans,
			err: err,
		}:
		case <-ctx.Done():
			return
		}

	}()
	return ch
}

func (s *ServerConfig) resolve(req *dns.Msg, resultCh chan<- *queryResult, ctx context.Context) {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	m := &dns.Msg{}
	m.SetReply(req)
	for _, q := range m.Question {
		wg.Add(1)
		go func(q dns.Question) {
			defer wg.Done()
			q.Name = rmHttp(q.Name)
			req := &dns.Msg{}
			req = req.SetQuestion(q.Name, q.Qtype)
			req.Compress = true
			dnsClient.Timeout = timeout
			ch := s.query(req, ctx)
			select {
			case res := <-ch:
				if res.err != nil {
					log.Errorf("Resolving %s error: %v\n", q.Name, res.err)
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
	if len(m.Answer) > 0 {
		resultCh <- &queryResult{s, m}
	}
}

func filter(results <-chan *queryResult) (*queryResult, error) {
	var cached *queryResult
	var confirmed bool
	var servers []string

	for res := range results {
		servers = append(servers, res.server.IP)
		ansIsCN, err := isChineseARecord(res.ans)
		if err != nil {
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

func queryServers(req *dns.Msg, servers []*ServerConfig) (*dns.Msg, error) {
	chResults := make(chan *queryResult, len(servers))
	go func() {
		<-time.After(timeout)
		close(chResults)
	}()

	ctx, cancel := context.WithCancel(context.Background())

	for _, server := range servers {
		go server.resolve(req, chResults, ctx)
	}

	result, err := filter(chResults)
	cancel()

	if err != nil {
		return &dns.Msg{}, fmt.Errorf("Resolving %s error: %v\n", req.Question[0].Name, err)
	}
	return result.ans, nil
}
