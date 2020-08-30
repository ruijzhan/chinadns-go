package main

import (
	"github.com/ruijzhan/chinadns-go/pkg/dns"
	"log"
)

func main() {
	err := dns.RunDNSServer()
	if err != nil {
		log.Fatal(err)
	}
}
