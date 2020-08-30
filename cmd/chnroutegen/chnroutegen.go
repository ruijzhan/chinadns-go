package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ruijzhan/chinadns-go/pkg/cidr"
	"log"
	"net/http"
	"os"
)

var (
	cidrs       cidr.CIDRs
	countryCode string
	outFile     string
	apnicURL    string
	toStdout    bool
)

func init() {
	flag.StringVar(&outFile, "w", "./chnroute.json", "Output file path")
	flag.StringVar(&countryCode, "c", "CN", "Country Code")
	flag.BoolVar(&toStdout, "o", false, "Write to stdout")
	flag.StringVar(&apnicURL, "apnoc", "https://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest", "CIDR url")
}

func main() {
	flag.Parse()

	log.Println("Reading from ", apnicURL)
	resp, err := http.Get(apnicURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	log.Println("Parsing ...")
	cidrs.List = cidr.CirdsByCountry(resp.Body, countryCode)
	content, err := json.Marshal(cidrs)
	if err != nil {
		log.Fatal(err)
	}

	if toStdout {
		_, err := fmt.Fprint(os.Stdout, string(content))
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	log.Println("Writing China list to file ", outFile)
	f, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write(content)
}
