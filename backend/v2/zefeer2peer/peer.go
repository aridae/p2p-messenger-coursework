package zefeer2peer

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net"
	"time"
)

type SharedKey struct {
	RemoteKey []byte
	LocalKey  []byte
	Secret    []byte
}

func (sk *SharedKey) Update(remoteKey []byte, localKey []byte) {
	log.Println("print Update shared key info")

	if remoteKey != nil {
		sk.RemoteKey = remoteKey
	}

	if localKey != nil {
		sk.LocalKey = localKey
	}

	if sk.RemoteKey != nil && sk.LocalKey != nil {
		secret := CalcSharedSecret(sk.RemoteKey, sk.LocalKey)
		sk.Secret = secret[:32]
	}
}

type Peer struct {
	Name      string
	PubKey    ed25519.PublicKey
	Conn      *net.Conn
	FirstSeen string
	LastSeen  string
	SharedKey SharedKey
	MESSGBUF  chan *MessageBuffed
}

func (p Peer) String() string {
	return string(p.Name) + ":" + hex.EncodeToString(p.PubKey)
}

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
		MESSGBUF: make(chan *MessageBuffed, 100),
	}
}

func (p *Peer) UpdatePeerOnZPING(envelope *Envelope) error {
	log.Println("print UpdatePeerOnZPING")
	if string(envelope.Cmd) != string(ZPING) {
		return errors.New("invalid command")
	}

	zping := &PeerZPING{}
	err := json.Unmarshal(envelope.Body, zping)
	if err != nil {
		return err
	}

	rawPubKey, err := hex.DecodeString(zping.PubKey)
	if err != nil {
		return err
	}

	rawExKey, err := hex.DecodeString(zping.ExKey)
	if err != nil {
		return err
	}
	p.Name = zping.Name
	p.PubKey = rawPubKey
	p.SharedKey.Update(rawExKey, nil)
	return nil
}
