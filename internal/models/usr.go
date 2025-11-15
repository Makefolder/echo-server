// Project: Echo Server
// Description: Self-hosted communication server.
// Designed to be easily customisable and allow custom client implementations.
// Author: Makefolder
// Copyright (C) 2025, Artemii Fedotov <artemii.fedotov@tutamail.com>

package models

import "github.com/google/uuid"

type Usr struct {
	UID      uuid.UUID
	Username string  `json:"username"`
	Prefix   *string `json:"prefix,omitempty"`
}
