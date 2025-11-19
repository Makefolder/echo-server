// Project: Echo Server
// Description: Self-hosted communication server.
// Designed to be easily customisable and allow custom client implementations.
// Author: Makefolder
// Copyright (C) 2025, Artemii Fedotov <artemii.fedotov@tutamail.com>

package serv

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/Makefolder/echo-server/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const bufSize int = 1024 * 1024 * 4 // 4MB

type Serialisable interface {
	Serialize() []byte
}

type (
	CmdHandler func(author models.Usr, args ...string) Serialisable
	Option     func(*EchoServ)

	Client struct {
		Usr  models.Usr
		Conn net.Conn
	}

	EchoServ struct {
		log  *zap.SugaredLogger
		host string
		port uint16

		cmds map[string]CmdHandler

		mu    sync.RWMutex
		conns map[uuid.UUID]Client

		vcmu sync.RWMutex
		vc   map[uuid.UUID]Client
	}
)

func WithCmd(cmd string, handler CmdHandler) Option {
	return func(e *EchoServ) {
		e.cmds[cmd] = handler
	}
}

func New(log *zap.SugaredLogger, host string, port uint16, opts ...Option) *EchoServ {
	serv := &EchoServ{
		log:  log,
		host: host,
		port: port,

		conns: make(map[uuid.UUID]Client),
		cmds:  make(map[string]CmdHandler, len(opts)),
	}

	for _, opt := range opts {
		opt(serv)
	}

	return serv
}

func (e *EchoServ) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", e.host, e.port))
	if err != nil {
		return fmt.Errorf("failed to start echo server instance: %w", err)
	}

	go func() {
		<-ctx.Done()
		e.log.Debug("shut down command received")
		listener.Close()
	}()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					e.log.Errorf("failed to accept connection: %v", err)
					continue
				}
			}

			go e.handleConn(conn)
		}
	}()

	return nil
}
