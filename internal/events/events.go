// Project: Echo Server
// Description: Self-hosted communication server.
// Designed to be easily customisable and allow custom client implementations.
// Author: Makefolder
// Copyright (C) 2025, Artemii Fedotov <artemii.fedotov@tutamail.com>

package events

import (
	"time"

	"github.com/Makefolder/echo-server/internal/models"
	"github.com/google/uuid"
)

const (
	ChanTypeVoice ChanType = "voice"
	ChanTypeText  ChanType = "text"
)

type (
	ChanType string

	Channel struct {
		UID  uuid.UUID `json:"uid"`
		Type ChanType  `json:"type"`
		Name string    `json:"name"`
	}

	ServEvent interface {
		Serialise() string
	}
)

// Common fields
type BaseEvent struct {
	Type string `json:"type"`
}

// Events

type Sys struct {
	BaseEvent
	Msg       string    `json:"msg"`
	Timestamp time.Time `json:"timestamp"`
}

type Broadcast struct {
	BaseEvent
	Msg       string     `json:"msg"`
	Timestamp time.Time  `json:"timestamp"`
	Usr       models.Usr `json:"usr"`
}

type Auth struct {
	BaseEvent
	Name   string  `json:"name"`
	Prefix *string `json:"prefix,omitempty"`
}

type Msg struct {
	BaseEvent
	Msg string `json:"msg"`
}

type Voice struct {
	BaseEvent
	Data []byte `json:"data"`
}

type Cmd struct {
	BaseEvent
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"`
}

type UsrList struct {
	BaseEvent
	Users     []models.Usr `json:"users"`
	Len       int          `json:"len"`
	Timestamp time.Time    `json:"timestamp"`
}

type ChanList struct {
	BaseEvent
	Channels  []Channel `json:"channels"`
	Len       int       `json:"len"`
	Timestamp time.Time `json:"timestamp"`
}
