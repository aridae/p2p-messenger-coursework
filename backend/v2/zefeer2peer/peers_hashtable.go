package zefeer2peer

import (
	"strings"
	"sync"
)

type HashKey string

// хэш таблица для хранения пиров
// наш клиент привязан к хранилищу пиров
// потому что без него не сможет обрабатывать запросы
type PeersHashTable struct {
	rwmux *sync.RWMutex
	peers map[HashKey]*Peer
}

func NewPeersHashTable() *PeersHashTable {
	return &PeersHashTable{
		rwmux: new(sync.RWMutex),
		peers: make(map[HashKey]*Peer),
	}
}

func (p PeersHashTable) Put(peer *Peer) {
	p.rwmux.Lock()
	defer p.rwmux.Unlock()

	p.peers[HashKey(peer.PubKey)] = peer
}

func (p PeersHashTable) Get(key HashKey) (*Peer, bool) {
	p.rwmux.RLock()
	defer p.rwmux.RUnlock()

	for k, v := range p.peers {
		if strings.EqualFold(string(k), string(key)) {
			return v, true
		}
	}
	return nil, false
}

func (p PeersHashTable) Remove(peer *Peer) (found bool) {
	p.rwmux.RLock()
	defer p.rwmux.RUnlock()

	(*peer.Conn).Close()
	delete(p.peers, HashKey(peer.PubKey))
	return
}

func (p PeersHashTable) Empty() {
	p.rwmux.RLock()
	defer p.rwmux.RUnlock()

	for k, v := range p.peers {
		(*v.Conn).Close()
		delete(p.peers, HashKey(k))
	}
}

func (p PeersHashTable) ToList() []PeerName {
	names := make([]PeerName, 0)
	for k, v := range p.peers {
		names = append(names, PeerName{
			PubKey: string(k),
			Name:   v.Name,
		})
	}
	return names
}
