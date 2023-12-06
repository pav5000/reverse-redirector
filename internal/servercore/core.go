package servercore

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/pav5000/reverse-redirector/internal/proto"
)

const (
	ConnectionCheckPeriod = time.Second * 20
)

type Core struct {
	lock          sync.Mutex
	clientHandler *ClientConnectionHandler
}

type ForwardRequest struct {
	DestAddr   string
	Connection net.Conn
}

func New() *Core {
	return &Core{}
}

func (c *Core) ListenClients(addr string, token string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("cannot listen to addr '%s': %v", addr, err)
	}
	log.Println("Listening for client connections at", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("cannot accept a connection: %v", err)
		}

		go c.HandleClientConnection(conn, token)

	}
}

var ErrIncorrectToken = errors.New("incorrect token")

type ClientConnectionHandler struct {
	Connection net.Conn
	EndChan    chan struct{}
}

func (c *Core) HandleClientConnection(conn net.Conn, token string) error {
	defer conn.Close()
	if !checkToken(conn, token) {
		return ErrIncorrectToken
	}

	err := proto.SendMsg(conn, "ok")
	if err != nil {
		return err
	}

	// Handshake done

	handler := &ClientConnectionHandler{
		Connection: conn,
		EndChan:    make(chan struct{}),
	}
	// Setting this handler as active handler which will process the next forwarding request
	c.ReplaceClientConnectionHandler(handler)

	// Waiting for a forwarding request
	<-handler.EndChan
	return nil
}

func (c *Core) ProcessForwardRequest(req ForwardRequest) {
	handler := c.RemoveConnectionHandler()
	if handler == nil {
		req.Connection.Close()
		return
	}
	handler.ProcessForwardRequest(req)
}

func (c *Core) ReplaceClientConnectionHandler(handler *ClientConnectionHandler) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.clientHandler != nil {
		c.clientHandler.Close()
	}

	c.clientHandler = handler
}

func (c *Core) RemoveConnectionHandler() *ClientConnectionHandler {
	var handler *ClientConnectionHandler
	c.lock.Lock()
	handler = c.clientHandler
	c.clientHandler = nil
	c.lock.Unlock()
	return handler
}

func checkToken(conn net.Conn, token string) bool {
	inMsg, err := proto.ReceiveMsg(conn)
	if err != nil {
		return false
	}
	return inMsg == token
}

func (h *ClientConnectionHandler) Close() {
	_ = h.Connection.Close()
	close(h.EndChan)
}

func (h *ClientConnectionHandler) ProcessForwardRequest(req ForwardRequest) error {
	defer h.Close()

	err := proto.SendDialRequest(h.Connection, req.DestAddr)
	if err != nil {
		return err
	}

	go func() {
		_, _ = io.Copy(req.Connection, h.Connection)
		req.Connection.Close()
		h.Connection.Close()
	}()

	_, _ = io.Copy(h.Connection, req.Connection)

	return nil
}
