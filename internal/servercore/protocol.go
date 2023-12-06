package servercore

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

const (
	MaxMessageSize = 1024
)

var (
	ErrMsgTooBig = errors.New(fmt.Sprintf("message must not exceed %d bytes", MaxMessageSize))
)

func SendMsg(w io.Writer, msg string) error {
	if len(msg) > MaxMessageSize {
		return ErrMsgTooBig
	}
	size := make([]byte, 2)
	binary.LittleEndian.PutUint16(size, uint16(len(msg)))
	_, err := w.Write(size)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msg))
	return err
}

func ReceiveMsg(w io.Reader) (string, error) {
	size := make([]byte, 2)
	_, err := w.Read(size)
	if err != nil {
		return "", err
	}
	parsedSize := binary.LittleEndian.Uint16(size)
	if parsedSize > MaxMessageSize {
		return "", ErrMsgTooBig
	}
	buf := make([]byte, parsedSize)
	_, err = io.ReadFull(w, buf)
	return string(buf), err
}

func SendDialRequest(conn io.ReadWriter, addr string) error {
	err := SendMsg(conn, "Dial "+addr)
	if err != nil {
		return errors.WithMessage(err, "sending dial message")
	}

	msg, err := ReceiveMsg(conn)
	if err != nil {
		return errors.WithMessage(err, "receiving dial confirmation")
	}

	if msg == "ok" {
		return nil
	}

	return errors.New("remote answered: " + msg)
}
