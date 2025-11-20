// Project: Echo Server
// Description: Self-hosted communication server.
// Designed to be easily customisable and allow custom client implementations.
// Author: Makefolder
// Copyright (C) 2025, Artemii Fedotov <artemii.fedotov@tutamail.com>

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

var (
	serverChannels = []events.Channel{
		{
			UID:  uuid.New(),
			Name: "general",
			Type: events.ChanTypeText,
		},
		{
			UID:  uuid.New(),
			Name: "voice",
			Type: events.ChanTypeVoice,
		},
	}
	chanLen = len(serverChannels)
)

func (e *EchoServ) handleConn(conn net.Conn) {
	// 1. Auth
	usr, err := e.authenticate(conn)
	if err != nil {
		e.SendSys(conn, fmt.Sprintf("unauthorized: %s", err.Error()))
		return
	}
	e.SendSys(conn, fmt.Sprintf("authorised as %s", usr.Username))

	e.mu.Lock()
	e.conns[usr.UID] = Client{
		Conn: conn,
		Usr:  usr,
	}
	e.mu.Unlock()

	defer func() {
		conn.Close()

		e.mu.Lock()
		delete(e.conns, usr.UID)
		e.mu.Unlock()
	}()

	// 2. Main conn loop
	for {
		event, err := e.readEvent(conn)
		if err != nil {
			if err == io.EOF {
				e.log.Debug("connection closed")
			} else {
				e.log.Errorf("failed to read event: %v", err)
			}
			break
		}

		// 3. Handle any user interactions
		if err := e.handleEvent(conn, event, usr); err != nil {
			e.SendSys(conn, err.Error())
		}
	}
}

func (e *EchoServ) handleEvent(conn net.Conn, event any, usr models.Usr) error {
	switch ev := event.(type) {
	case events.Msg:
		e.log.Debugf("msg event detected: %v", ev)
		if err := e.Broadcast(ev, usr); err != nil {
			return fmt.Errorf("failed to broadcast msg event: %w", err)
		}
	case events.Cmd:
		e.log.Debugf("cmd event detected: %v", ev)
		if err := e.sendCmd(conn, ev, usr); err != nil {
			return fmt.Errorf("failed to send cmd event: %w", err)
		}
	default:
		e.log.Debug("other event detected")
	}
	return nil
}

func (e *EchoServ) sendCmd(conn net.Conn, event events.Cmd, author models.Usr) error {
	switch event.Cmd {
	case "voice":
		_, ok := e.vc[author.UID]
		if ok {
			return errors.New("user is already in voice chat")
		}
		go e.HandleVc(conn, author)
		return nil
	case "users":
		return e.SendUsrList(conn)
	case "channels":
		return e.SendChanList(conn)
	default:
		return e.customCmd(conn, event, author)
	}
}

func (e *EchoServ) customCmd(conn net.Conn, event events.Cmd, author models.Usr) error {
	handler, ok := e.cmds[event.Cmd]
	if !ok {
		return ErrCmdNotFound
	}

	servCtx := e.newCtx(Client{
		Usr:  author,
		Conn: conn,
	})

	handler(servCtx, event.Args...)
	return nil
}

func (e *EchoServ) SendUsrList(conn net.Conn) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	usrList := events.UsrList{
		BaseEvent: events.BaseEvent{
			Type: "usr_list",
		},
	}
	usrList.Users = make([]models.Usr, len(e.conns))
	usrList.Len = len(e.conns)

	for _, cli := range e.conns {
		usrList.Users = append(usrList.Users, cli.Usr)
	}

	usrList.Timestamp = time.Now().UTC()
	n, err := fmt.Fprintf(conn, "%s\n\r", usrList.Serialise())
	if err != nil {
		return fmt.Errorf("failed to send chan list event: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("zero bytes sent: %w", err)
	}

	return nil
}

func (e *EchoServ) SendChanList(conn net.Conn) error {
	chanList := events.ChanList{
		BaseEvent: events.BaseEvent{
			Type: "chan_list",
		},
		Channels:  serverChannels,
		Len:       chanLen,
		Timestamp: time.Now().UTC(),
	}

	n, err := fmt.Fprintf(conn, "%s\n\r", chanList.Serialise())
	if err != nil {
		return fmt.Errorf("failed to send chan list event: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("zero bytes sent: %w", err)
	}

	return nil
}

func (e *EchoServ) HandleVc(conn net.Conn, author models.Usr) {
	e.vcmu.Lock()
	e.vc[author.UID] = Client{
		Usr:  author,
		Conn: conn,
	}
	e.vcmu.Unlock()

	defer func() {
		e.vcmu.Lock()
		delete(e.vc, author.UID)
		e.vcmu.Unlock()
	}()

	for {
		event, err := e.readEvent(conn)
		if err != nil {
			if err == io.EOF {
				e.log.Debug("connection closed")
			} else {
				e.log.Errorf("failed to read event: %v", err)
			}
			break
		}

		switch ev := event.(type) {
		case events.Voice:
			e.log.Debugf("voice event detected: %v", ev)
			for _, cli := range e.vc {
				n, err := fmt.Fprintf(cli.Conn, "%s\n\r", ev.Serialise())
				if err != nil {
					e.log.Error("failed to broadcast voice event: %v", err)
					break
				}
				if n == 0 {
					break
				}
				e.log.Debugf("sent voice event: %d (bytes)", n)
			}
		default:
			e.log.Debug("other event detected in voice chat handling")
		}
	}
}
