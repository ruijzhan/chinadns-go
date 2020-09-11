package dns

import (
	"fmt"
	"github.com/miekg/dns"
	"github.com/ruijzhan/chinadns-go/pkg/options"
	"github.com/ruijzhan/country-cidr"
	"net/url"
	"strings"
)

var (
	opts *options.Options
)

func init() {
	opts = options.NewOptions()
}

func isChineseARecord(msg *dns.Msg) (bool, error) {
	for _, rr := range msg.Answer {
		if rr, ok := rr.(*dns.A); ok {
			cn := country_cidr.Country("CN").Contains(rr.A)
			return cn, nil
		}
	}
	if msg.Question[0].Qtype == dns.TypeAAAA {
		return false, nil
	}
	return false, fmt.Errorf("query to %s returns %d ANS, %d AUTH, %d EXT",
		msg.Question[0].Name, len(msg.Answer), len(msg.Ns), len(msg.Extra))
}

func isChineseIP(ip string) bool {
	return country_cidr.Country("CN").ContainsIPstr(ip)
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

func getIP(msg *dns.Msg) string {
	for _, rr := range msg.Answer {
		if rr, ok := rr.(*dns.A); ok {
			return rr.A.String()
		}
	}
	return ""
}
