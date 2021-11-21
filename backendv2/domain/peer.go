package domain

import (
	"crypto/ed25519"
	"net"
	"time"
)

// TODO: прогуглить, что значит каждое поле
type Peer struct {
	PubKey    ed25519.PublicKey
	Conn      *net.Conn
	Name      string
	FirstSeen string
	LastSeen  string
	SharedKey SharedKey
}

type PeerName struct {
	Name   string
	PubKey string
}

//NewPeer create new peer struct by socket connection
func NewPeer(conn net.Conn) *Peer {
	return &Peer{
		PubKey:    nil,
		Conn:      &conn,
		Name:      conn.RemoteAddr().String(),
		FirstSeen: time.Now().String(),
		LastSeen:  time.Now().String(),
		SharedKey: SharedKey{
			RemoteKey: nil,
			LocalKey:  nil,
			Secret:    nil,
		},
	}
}
