package dns

import (
	"encoding/json"
	"fmt"
	"github.com/miekg/dns"
	"github.com/ruijzhan/chinadns-go/pkg/cidr"
	"github.com/ruijzhan/chinadns-go/pkg/options"
	"github.com/yl2chen/cidranger"
	"io/ioutil"
	"log"
	"net"
	"os"
)

var (
	chnRanger cidranger.Ranger
	opts      *options.Options
)


func init()  {
	opts = options.NewOptions()

	f, err := os.Open(opts.CHNRoutePath)
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

func isChineseIP(ip string) (bool,error) {
	return chnRanger.Contains(net.ParseIP(ip))
}
