package proto

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

//SharedKey ECDHE shared key
type SharedKey struct {
	RemoteKey []byte
	LocalKey  []byte
	Secret    []byte
}

//Update shared key info
func (sk *SharedKey) Update(remoteKey []byte, localKey []byte) {
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

//Peer basic struct to describe peer
type Peer struct {
	PubKey    ed25519.PublicKey
	Conn      *net.Conn
	Name      string
	FirstSeen string
	LastSeen  string
	Peers     *Peers
	SharedKey SharedKey
}

func (p Peer) String() string {
	return string(p.Name) + ":" + hex.EncodeToString(p.PubKey)
}

//NewPeer create new peer struct by socket connection
func NewPeer(conn net.Conn) *Peer {
	return &Peer{
		PubKey:    nil,
		Conn:      &conn,
		Name:      conn.RemoteAddr().String(),
		FirstSeen: time.Now().String(),
		LastSeen:  time.Now().String(),
		Peers:     NewPeers(),
		SharedKey: SharedKey{
			RemoteKey: nil,
			LocalKey:  nil,
			Secret:    nil,
		},
	}
}

//UpdatePeer Update peer struct after handshake
func (p *Peer) UpdatePeer(envelope *Envelope) error {
	if string(envelope.Cmd) != "HANDP" {
		return errors.New("invalid command")
	}

	handShake := &HandShake{}
	err := json.Unmarshal(envelope.Content, handShake)
	if err != nil {
		return err
	}

	rawPubKey, err := hex.DecodeString(handShake.PubKey)
	if err != nil {
		return err
	}

	rawExKey, err := hex.DecodeString(handShake.ExKey)
	if err != nil {
		return err
	}

	// TODO: проверить подпись

	p.Name = handShake.Name
	p.PubKey = rawPubKey

	p.SharedKey.Update(rawExKey, nil)
	return nil
}

//Peers synchronised list of peers
// Явно указано, что пирс реализует
type Peers struct {
	rwmux *sync.RWMutex
	peers map[string]*Peer
}

//NewPeers create new list of peers
func NewPeers() *Peers {
	return &Peers{
		rwmux: new(sync.RWMutex),
		peers: make(map[string]*Peer),
	}
}

//Put put new peer to list
func (p Peers) Put(peer *Peer) {
	p.rwmux.Lock()
	defer p.rwmux.Unlock()

	p.peers[string(peer.PubKey)] = peer
}

//Get find and get peer in list
func (p Peers) Get(key string) (peer *Peer, found bool) {
	p.rwmux.RLock()
	defer p.rwmux.RUnlock()

	peer, found = p.peers[key]
	return
}

//Remove remove peer from list
func (p Peers) Remove(peer *Peer) (found bool) {
	p.rwmux.RLock()
	defer p.rwmux.RUnlock()

	delete(p.peers, string(peer.PubKey))
	return
}

//PeerList return json list of peers
func (p Peers) PeerList() *WsPeerList {

	peerList := &WsPeerList{
		WsCmd: WsCmd{
			Cmd: "PEERS",
		},
		Peers: []PeerName{},
	}

	log.Printf("total peers: %v", len(p.peers))

	for _, el := range p.peers {
		peerList.Peers = append(peerList.Peers, PeerName{
			Name:   el.Name,
			PubKey: hex.EncodeToString(el.PubKey),
		})
	}

	return peerList
}
