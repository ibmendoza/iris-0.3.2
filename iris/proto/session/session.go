// Iris - Decentralized cloud messaging
// Copyright (c) 2013 Project Iris. All rights reserved.
//
// Iris is dual licensed: you can redistribute it and/or modify it under the
// terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
//
// The framework is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// Alternatively, the Iris framework may be used in accordance with the terms
// and conditions contained in a signed written agreement between you and the
// author(s).

// Package session implements an encrypted data stream, authenticated through
// the station-to-station key exchange.
package session

import (
	"hash"
	"io"

	//"code.google.com/p/go.crypto/hkdf"
	"github.com/project-iris/iris/config"
	"github.com/project-iris/iris/proto/link"
	"github.com/project-iris/iris/proto/stream"
	"golang.org/x/crypto/hkdf"
)

// Accomplishes secure and authenticated full duplex communication.
type Session struct {
	kdf io.Reader // Key derivation function to expand the master key

	CtrlLink *link.Link // Network connection for high priority control messages
	DataLink *link.Link // Network connection for low priority data messages
}

// Creates a new, double link session for authenticated data transfer. The
// initiator is used to decide the key derivation order for the channels.
func newSession(conn *stream.Stream, secret []byte, server bool) *Session {
	// Create the key derivation function
	hasher := func() hash.Hash { return config.HkdfHash.New() }
	hkdf := hkdf.New(hasher, secret, config.HkdfSalt, config.HkdfInfo)

	// Create the encrypted control link
	return &Session{
		kdf:      hkdf,
		CtrlLink: link.New(conn, hkdf, server),
	}
}

// Finalizes a session by creating the secondary data link.
func (s *Session) init(conn *stream.Stream, server bool) {
	s.DataLink = link.New(conn, s.kdf, server)
}

// Starts the session data transfers on the control and data channels.
func (s *Session) Start(cap int) {
	s.CtrlLink.Start(cap)
	s.DataLink.Start(cap)
}

// Terminates the data transfers on the two channels
func (s *Session) Close() error {
	res := s.CtrlLink.Close()
	if err := s.DataLink.Close(); res == nil {
		res = err
	}
	return res
}
