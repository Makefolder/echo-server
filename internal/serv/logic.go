package serv

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Makefolder/echo-server/internal/events"
	"github.com/Makefolder/echo-server/internal/models"
	"github.com/google/uuid"
)

func (e *EchoServ) sendSys(conn net.Conn, msg string) error {
	sysEvent := events.Sys{
		BaseEvent: events.BaseEvent{
			Type: "sys",
		},
		Msg:       msg,
		Timestamp: time.Now().UTC(),
	}

	n, err := fmt.Fprintf(conn, "%s\n\r", sysEvent.Serialise())
	if err != nil {
		return fmt.Errorf("failed to send sys event: %w", err)
	}

	if n == 0 {
		return fmt.Errorf("failed to send sys event: %w", io.EOF)
	}

	e.log.Debugf("sent sys event: %d (bytes)", n)
	return nil
}

func (e *EchoServ) broadcast(msgEvent events.Msg, author models.Usr) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	broadcastEvent := events.Broadcast{
		BaseEvent: events.BaseEvent{
			Type: "broadcast",
		},
		Msg:       msgEvent.Msg,
		Timestamp: time.Now().UTC(),
		Usr:       author,
	}

	for _, cli := range e.conns {
		n, err := fmt.Fprintf(cli.Conn, "%s\n\r", broadcastEvent.Serialise())
		if err != nil {
			return fmt.Errorf("failed to broadcast msg event: %w", err)
		}
		if n == 0 {
			return fmt.Errorf("zero bytes sent: %w", err)
		}
		e.log.Debugf("sent msg event: %d (bytes)", n)
	}
	return nil
}

func (e *EchoServ) authenticate(conn net.Conn) (models.Usr, error) {
	var usr models.Usr
	event, err := e.readEvent(conn)
	if err != nil {
		if err == io.EOF {
			e.log.Debug("connection closed")
			return usr, err
		}
		e.log.Errorf("failed to read event: %v", err)
		return usr, fmt.Errorf("failed to read event: %w", err)
	}

	switch ev := event.(type) {
	case events.Auth:
		e.log.Debugf("auth event detected: %v", ev)
		usr.UID = uuid.New()
		usr.Username = ev.Name
		usr.Prefix = ev.Prefix
	default:
		e.log.Debug("in authenticate: no auth event provided")
		return usr, errors.New("no auth event provided")
	}

	return usr, nil
}

func (e *EchoServ) readEvent(conn net.Conn) (any, error) {
	data := make([]byte, bufSize)
	n, err := conn.Read(data)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, io.EOF
	}

	e.log.Debugf("reading event: %d (bytes): %s", n, data[:n])
	event, err := events.Parse(data[:n])
	if err != nil {
		e.sendSys(conn, err.Error())
		return nil, err
	}

	return event, nil
}
