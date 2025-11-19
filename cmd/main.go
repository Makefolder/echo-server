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

type PingRes struct {
	Author models.Usr `json:"author"`
	Msg    string     `json:"msg"`
	Args   []string   `json:"args"`
}

func (p PingRes) Serialize() []byte {
	res, _ := json.Marshal(p)
	return res
}

func handlePing(author models.Usr, args ...string) serv.Serialisable {
	res := PingRes{
		Author: author,
		Msg:    "pong",
		Args:   args,
	}

	return res
}

func main() {
	logger, err := zap.NewDevelopment(zap.IncreaseLevel(lvl))
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	slog := logger.Sugar()
	defer slog.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := serv.New(slog, host, port,
		serv.WithCmd("ping", handlePing),
	)

	if err := e.Start(ctx); err != nil {
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
