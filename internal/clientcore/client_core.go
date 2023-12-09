package clientcore

import (
	"io"
	"net"

	"github.com/pav5000/reverse-redirector/internal/proto"
	"github.com/pkg/errors"
)

var (
	ErrCannotGetServerConfirmation = errors.New("cannot get token confirmation from server")
	ErrReceivingDialRequest        = errors.New("error while receiving dial request")
)

type ClientCore struct {
	token string
}

func New(token string) *ClientCore {
	return &ClientCore{
		token: token,
	}
}

type Connection struct {
	serverConn net.Conn
	dialAddr   string
}

func (c *ClientCore) GetServerConnection(serverAddr string) (*Connection, error) {
	serverConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	err = proto.SendMsg(serverConn, c.token)
	if err != nil {
		return nil, err
	}

	err = proto.ReceiveOkOrError(serverConn)
	if err != nil {
		return nil, ErrCannotGetServerConfirmation
	}

	return &Connection{
		serverConn: serverConn,
	}, nil
}

func (c *Connection) WaitForTask() error {
	dialAddr, err := proto.ReceiveDialRequest(c.serverConn)
	if err != nil {
		return err
	}
	c.dialAddr = dialAddr
	return nil
}

func (c *Connection) ProcessTask() error {
	defer c.serverConn.Close()

	redirectConn, err := net.Dial("tcp", c.dialAddr)
	if err != nil {
		proto.SendError(c.serverConn, err.Error())
		return err
	}
	defer redirectConn.Close()

	err = proto.SendOk(c.serverConn)
	if err != nil {
		return err
	}

	go func() {
		_, _ = io.Copy(c.serverConn, redirectConn)
		c.serverConn.Close()
		redirectConn.Close()
	}()

	_, _ = io.Copy(redirectConn, c.serverConn)
	return nil
}
