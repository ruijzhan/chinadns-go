package chinadns

import (
	"context"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
)

type remoteDNS struct {
	IP        string
	Port      int
	IsCN      bool
	dnsClient *dns.Client
}

func (svr *remoteDNS) String() string {
	return svr.IP + ":" + strconv.Itoa(svr.Port)
}

func (svr *remoteDNS) query(req *dns.Msg, ctx context.Context) <-chan *queryResult {
	ch := make(chan *queryResult)

	go func() {
		ans, _, err := svr.dnsClient.Exchange(req, svr.String())
		select {
		case ch <- &queryResult{
			server: svr,
			answer: ans,
			err:    err,
		}:
		case <-ctx.Done():
			return
		}

	}()
	return ch
}

func (svr *remoteDNS) resolve(req *dns.Msg, resultCh chan<- *queryResult, ctx context.Context) {
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
			ch := svr.query(req, ctx)
			select {
			case res := <-ch:
				if res.err != nil {
					log.Errorf("Resolving %svr error: %v\n", q.Name, res.err)
				} else {
					mu.Lock()
					m.Answer = append(m.Answer, res.answer.Answer...)
					m.Ns = append(m.Ns, res.answer.Ns...)
					m.Extra = append(m.Extra, res.answer.Extra...)
					mu.Unlock()
				}
			case <-ctx.Done():
				return
			}
		}(q)
	}
	wg.Wait()
	if len(m.Answer) > 0 {
		resultCh <- &queryResult{svr, m, nil}
	}
}
