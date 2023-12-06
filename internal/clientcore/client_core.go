package clientcore

import (
	"errors"
	"io"
	"net"

	"github.com/pav5000/reverse-redirector/internal/proto"
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

func (c *ClientCore) ProcessTaskFromServer(serverAddr string) error {
	serverConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}
	defer serverConn.Close()

	err = proto.SendMsg(serverConn, c.token)
	if err != nil {
		return err
	}

	err = proto.ReceiveOkOrError(serverConn)
	if err != nil {
		return ErrCannotGetServerConfirmation
	}

	dialAddr, err := proto.ReceiveDialRequest(serverConn)
	if err != nil {
		return ErrReceivingDialRequest
	}

	redirectConn, err := net.Dial("tcp", dialAddr)
	if err != nil {
		proto.SendError(serverConn, err.Error())
		return err
	}
	defer redirectConn.Close()

	err = proto.SendOk(serverConn)
	if err != nil {
		return err
	}

	go func() {
		_, _ = io.Copy(serverConn, redirectConn)
		serverConn.Close()
		redirectConn.Close()
	}()

	_, _ = io.Copy(redirectConn, serverConn)
	return nil
}
