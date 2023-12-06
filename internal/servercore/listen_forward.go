package servercore

import (
	"fmt"
	"log"
	"net"
)

func (c *Core) ListenForward(addrSrc, addrDst string) error {
	listener, err := net.Listen("tcp", addrSrc)
	if err != nil {
		return fmt.Errorf("cannot listen to addr '%s': %v", addrSrc, err)
	}
	log.Println("Listening forward", addrSrc, "->", addrDst)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("cannot accept a connection: %v", err)
		}

		go c.ProcessForwardRequest(ForwardRequest{
			Connection: conn,
			DestAddr:   addrDst,
		})
	}
}
