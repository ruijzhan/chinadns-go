package main

import (
	"github.com/ruijzhan/chinadns-go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	//Exit on signal
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Infof("Received signal: %v\n", sig.String())
		os.Exit(0)
	}()

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
}

func main() {
	err := chinadns.RunDNSServer()
	if err != nil {
		log.Fatal(err)
	}
}
