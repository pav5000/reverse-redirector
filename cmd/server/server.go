package main

import (
	"log"
	"time"

	"github.com/pav5000/reverse-redirector/internal/servercore"
	"github.com/pav5000/reverse-redirector/internal/utils"
)

const (
	ListenRetry = time.Second * 5
)

type Config struct {
	ListenAddr string     `yaml:"listen"`    // listen for client part connections
	Token      string     `yaml:"token"`     // token which client knows, it's not authentication, just to prevent easy abusing
	Redirects  []Redirect `yaml:"redirects"` // list of port redirects
}

type Redirect struct {
	From string `yaml:"from"` // listen address on server side
	To   string `yaml:"to"`   // connect address on client side
}

func main() {
	var conf Config
	utils.MustReadConfig("server.yml", &conf)

	if conf.ListenAddr == "" {
		log.Fatal("listen address shouldn't be empty")
	}
	if conf.Token == "" {
		log.Fatal("token shouldn't be empty")
	}

	core := servercore.New()

	for _, redirectConf := range conf.Redirects {
		redirectConf := redirectConf
		go func() {
			for {
				err := core.ListenForward(redirectConf.From, redirectConf.To)
				if err != nil {
					log.Println("Cannot listen redirect", redirectConf.From, "->", redirectConf.To)
					log.Println("will retry to listen in", ListenRetry)
					time.Sleep(ListenRetry)
					continue
				}
				log.Println("Redirect exited:", redirectConf.From, "->", redirectConf.To)
			}
		}()
	}

	for {
		err := core.ListenClients(conf.ListenAddr, conf.Token)
		if err != nil {
			log.Println("Listening clients at", conf.ListenAddr, "failed:", err)
		}
		log.Println("Will retry to listen in", ListenRetry)
		time.Sleep(ListenRetry)
	}
}
