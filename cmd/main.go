// Project: Echo Server
// Description: Self-hosted communication server.
// Designed to be easily customisable and allow custom client implementations.
// Author: Makefolder
// Copyright (C) 2025, Artemii Fedotov <artemii.fedotov@tutamail.com>

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Makefolder/echo-server/internal/httpclient"
	"github.com/Makefolder/echo-server/internal/models"
	"github.com/Makefolder/echo-server/internal/serv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	host string        = "0.0.0.0"
	port uint16        = 2001
	lvl  zapcore.Level = zap.DebugLevel
)

func handlePing(ctx *serv.Ctx, args ...string) {
	var res struct {
		Author models.Usr `json:"author"`
		Msg    string     `json:"msg"`
		Args   []string   `json:"args"`
	}

	res.Author = ctx.Cli.Usr
	res.Msg = "pong"
	res.Args = args

	serialised, err := json.Marshal(res)
	if err != nil {
		return
	}

	if err := ctx.Serv.SendSys(ctx.Cli.Conn, string(serialised)); err != nil {
		ctx.Log.Errorf("failed to send cmd event: %v", err)
	}
}

func main() {
	logger, err := zap.NewProduction(zap.IncreaseLevel(lvl))
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	slog := logger.Sugar()
	defer slog.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	params := serv.SetupParams{
		Log:    slog,
		Host:   host,
		Port:   port,
		WebCli: httpclient.New(nil, nil, 5*time.Second),
	}

	serv, err := serv.New(params,
		serv.WithCmd("ping", handlePing),
	)

	if err != nil {
		slog.Fatalf("failed to create echo server instance: %v", err)
	}

	if err := serv.Start(ctx); err != nil {
		slog.Fatalf("failed to start echo server instance: %v", err)
	}

	slog.Infof("echo server instance is up and running on %s:%d", host, port)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	slog.Info("shutting down echo server instance...")
	cancel()

	// Give it a moment to clean up
	time.Sleep(100 * time.Millisecond)
	slog.Info("echo instance is shut down. Thanks for using echo server!")
}
