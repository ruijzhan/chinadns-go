package chinadns

import (
	"fmt"
	"github.com/miekg/dns"
	cidr "github.com/ruijzhan/country-cidr"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strconv"
	"strings"
)

func isChineseARecord(msg *dns.Msg) (bool, error) {
	for _, rr := range msg.Answer {
		if rr, ok := rr.(*dns.A); ok {
			cn := cidr.Country("CN").Contains(rr.A)
			return cn, nil
		}
	}
	if msg.Question[0].Qtype == dns.TypeAAAA {
		return false, nil
	}
	return false, fmt.Errorf("query to %s returns %d ANS, %d AUTH, %d EXT",
		msg.Question[0].Name, len(msg.Answer), len(msg.Ns), len(msg.Extra))
}

func rmHttp(s string) string {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		u, err := url.Parse(s)
		if err != nil {
			return s
		}
		return u.Host + "."
	}
	return s
}

func parseServers(s string) []*ServerConfig {
	servers := make([]*ServerConfig, 0)
	for _, serverStr := range strings.Split(s, ",") {
		ts := strings.Split(serverStr, ":")
		isCN := cidr.Country("CN").ContainsIPstr(ts[0])
		if len(ts) == 1 {
			servers = append(servers, &ServerConfig{IP: ts[0], Port: 53, IsCN: isCN})
		} else if len(ts) == 2 {
			port, err := strconv.Atoi(ts[1])
			if err != nil {
				log.Fatalf("Invalid port number: %v", err)
			}
			servers = append(servers, &ServerConfig{IP: ts[0], Port: port, IsCN: isCN})
		} else {
			log.Fatalf("Invalid server address: %s", s)
		}
	}
	return servers
}
