// Project: Echo Server
// Description: Self-hosted communication server.
// Designed to be easily customisable and allow custom client implementations.
// Author: Makefolder
// Copyright (C) 2025, Artemii Fedotov <artemii.fedotov@tutamail.com>

package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Makefolder/echo-server/internal/models"
)

type ServEvent interface {
	Serialise() string
}

// Common fields
type BaseEvent struct {
	Type string `json:"type"`
}

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

func Parse(data []byte) (any, error) {
	var base BaseEvent
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.Type {
	case "sys":
		var e Sys
		err := json.Unmarshal(data, &e)
		return e, err
	case "broadcast":
		var e Broadcast
		err := json.Unmarshal(data, &e)
		return e, err
	case "auth":
		var e Auth
		err := json.Unmarshal(data, &e)
		return e, err
	case "msg":
		var e Msg
		err := json.Unmarshal(data, &e)
		return e, err
	case "voice":
		var e Voice
		err := json.Unmarshal(data, &e)
		return e, err
	case "cmd":
		var e Cmd
		err := json.Unmarshal(data, &e)
		return e, err
	case "usr_list":
		var e UsrList
		err := json.Unmarshal(data, &e)
		return e, err
	default:
		return nil, fmt.Errorf("unknown event type: %s", base.Type)
	}
}

func panicMsg(e any, err error) string {
	return fmt.Sprintf("failed to serialise event\nevent: %+v\nerror: %v", e, err)
}

func (e Sys) Serialise() string {
	data, err := json.Marshal(e)
	if err != nil {
		panic(panicMsg(e, err))
	}
	return string(data)
}

func (e Broadcast) Serialise() string {
	data, err := json.Marshal(e)
	if err != nil {
		panic(panicMsg(e, err))
	}
	return string(data)
}

func (e Voice) Serialise() string {
	data, err := json.Marshal(e)
	if err != nil {
		panic(panicMsg(e, err))
	}
	return string(data)
}

func (e UsrList) Serialise() string {
	data, err := json.Marshal(e)
	if err != nil {
		panic(panicMsg(e, err))
	}
	return string(data)
}
