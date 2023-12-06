package main

import (
	"log"
	"time"

	"github.com/pav5000/reverse-redirector/internal/clientcore"
	"github.com/pav5000/reverse-redirector/internal/utils"
)

const (
	ErrRetryTimeout = time.Second * 10
)

type Config struct {
	ServerAddrs []string `yaml:"servers"` // array of possible servers host:port (multiple servers for redundancy)
	Token       string   `yaml:"token"`   // it's not authentication, just to prevent easy abusing
}

func main() {
	var conf Config
	utils.MustReadConfig("client.yml", &conf)

	if len(conf.ServerAddrs) == 0 {
		log.Fatal("there should be at least one server address")
	}
	if conf.Token == "" {
		log.Fatal("token shouldn't be empty")
	}

	core := clientcore.New(conf.Token)

	for _, serverAddr := range conf.ServerAddrs {
		serverAddr := serverAddr
		go func() {
			for {
				err := core.ProcessTaskFromServer(serverAddr)
				if err != nil {
					log.Println("Error processing server", serverAddr, ":", err)
					time.Sleep(ErrRetryTimeout)
				}
			}
		}()
	}
	select {}
}
