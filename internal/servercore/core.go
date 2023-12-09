package servercore

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/pav5000/reverse-redirector/internal/proto"
)

const (
	ConnectionCheckPeriod                 = time.Second * 20
	WaitingForFreeClientConnectionTimeout = time.Second * 10
)

type Core struct {
	newClientChecked   chan struct{}        // we got a message here when the new client successfully connects
	checkedConnections chan net.Conn        // client connections with checked token
	forwardRequests    chan *ForwardRequest // incoming requests to forward connection to client
}

type ForwardRequest struct {
	DestAddr   string
	Connection net.Conn
}

func New() *Core {
	return &Core{
		newClientChecked:   make(chan struct{}),
		checkedConnections: make(chan net.Conn),
		forwardRequests:    make(chan *ForwardRequest),
	}
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

		go func(conn net.Conn) {
			if !checkToken(conn, token) {
				conn.Close()
				return
			}

			// Maintaning only one client connected at the same time
			select {
			case c.newClientChecked <- struct{}{}:
			default:
			}

			select {
			case c.checkedConnections <- conn:
			case <-c.newClientChecked:
				conn.Close()
				return
			}
		}(conn)
	}
}

var ErrIncorrectToken = errors.New("incorrect token")

func (c *Core) ProcessForwardRequest(req ForwardRequest) {
	defer req.Connection.Close()

	var clientConn net.Conn
	select {
	case clientConn = <-c.checkedConnections:
	case <-time.NewTimer(WaitingForFreeClientConnectionTimeout).C:
		return
	}
	defer clientConn.Close()

	err := proto.SendDialRequest(clientConn, req.DestAddr)
	if err != nil {
		log.Println("Error sending dial request:", err)
		return
	}

	go func() {
		_, _ = io.Copy(req.Connection, clientConn)
		req.Connection.Close()
		clientConn.Close()
	}()

	_, _ = io.Copy(clientConn, req.Connection)
}

func checkToken(conn net.Conn, token string) bool {
	inMsg, err := proto.ReceiveMsg(conn)
	if err != nil {
		return false
	}
	if inMsg != token {
		return false
	}

	err = proto.SendOk(conn)
	if err != nil {
		return false
	}
	return true
}
