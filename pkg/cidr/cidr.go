package cidr

import (
	"encoding/json"
	"github.com/yl2chen/cidranger"
	"io"
	"io/ioutil"
	"log"
	"net"
)

type CIDRs struct {
	List []string `json:"list"`
}

func loadCIDRs(r io.Reader) (cidrs CIDRs) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(b, &cidrs)
	return
}

func GetCHNRanger(cidrs *CIDRs) cidranger.Ranger {
	c := cidranger.NewPCTrieRanger()

	for _, cidr := range cidrs.List {
		_, nw, _ := net.ParseCIDR(cidr)
		c.Insert(cidranger.NewBasicRangerEntry(*nw))
	}
	return c
}
