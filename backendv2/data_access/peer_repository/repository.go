package peerrepository

import (
	"github.com/aridae/p2p-messenger-coursework/backendv2/domain"
)

type Repository interface {
	ListPeers() ([]domain.Peer, error)
	GetPeer(pubKey string) (*domain.Peer, error)
	RemovePeer(pubKey string) error
	AddPeer(*domain.Peer) (string, error)
}
