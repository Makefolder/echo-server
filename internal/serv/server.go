// Project: Echo Server
// Description: Self-hosted communication server.
// Designed to be easily customisable and allow custom client implementations.
// Author: Makefolder
// Copyright (C) 2025, Artemii Fedotov <artemii.fedotov@tutamail.com>

package serv

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/Makefolder/echo-server/internal/httpclient"
	"github.com/Makefolder/echo-server/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const bufSize int = 1024 * 1024 * 4 // 4MB

var (
	ErrNilLogger   = errors.New("nil logger")
	ErrNilWebCli   = errors.New("nil web client")
	ErrInvalidHost = errors.New("invalid host")
	ErrInvalidPort = errors.New("invalid port")
	ErrCmdNotFound = errors.New("command not found")
)

type (
	Ctx struct {
		Log  *zap.SugaredLogger
		Serv *EchoServ
		Web  *httpclient.HttpClient
		Cli  Client
	}

	CmdHandler func(ctx *Ctx, args ...string)
	Option     func(*EchoServ)

	SetupParams struct {
		Log    *zap.SugaredLogger
		Host   string
		Port   uint16
		WebCli *httpclient.HttpClient
	}

	Client struct {
		Usr  models.Usr
		Conn net.Conn
	}

	EchoServ struct {
		log *zap.SugaredLogger

		host string
		port uint16

		cmds   map[string]CmdHandler
		webCli *httpclient.HttpClient

		// regular chat (only one room for now)
		mu    sync.RWMutex
		conns map[uuid.UUID]Client

		// voice chat (only one room for now)
		vcmu sync.RWMutex
		vc   map[uuid.UUID]Client
	}
)

func WithCmd(cmd string, handler CmdHandler) Option {
	return func(e *EchoServ) {
		e.cmds[cmd] = handler
	}
}

func New(params SetupParams, opts ...Option) (*EchoServ, error) {
	if params.Log == nil {
		return nil, ErrNilLogger
	}

	if params.WebCli == nil {
		return nil, ErrNilWebCli
	}

	if params.Host == "" {
		return nil, ErrInvalidHost
	}

	if params.Port == 0 {
		return nil, ErrInvalidPort
	}

	serv := &EchoServ{
		log:    params.Log,
		host:   params.Host,
		port:   params.Port,
		webCli: params.WebCli,

		cmds:  make(map[string]CmdHandler, len(opts)),
		conns: make(map[uuid.UUID]Client),
		vc:    make(map[uuid.UUID]Client),
	}

	for _, opt := range opts {
		opt(serv)
	}

	return serv, nil
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
