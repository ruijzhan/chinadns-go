package chinadns

import (
	"fmt"
	"github.com/miekg/dns"
	cidr "github.com/ruijzhan/country-cidr"
	"net/url"
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
