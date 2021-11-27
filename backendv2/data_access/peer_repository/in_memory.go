package peerrepository

import (
	"github.com/aridae/p2p-messenger-coursework/backendv2/db"
	"github.com/aridae/p2p-messenger-coursework/backendv2/domain"
)

type InMemoryRepository struct {
	client *db.InMemoryClient
}

func (r *InMemoryRepository) ListPeers() ([]domain.Peer, error) {
	return nil, nil
}

func (r *InMemoryRepository) GetPeer(pubKey string) (*domain.Peer, error) {
	return nil, nil
}

func (r *InMemoryRepository) RemovePeer(pubKey string) error {
	return nil
}

func (r *InMemoryRepository) AddPeer(*domain.Peer) (string, error) {
	return "testKey", nil
}

func NewInMemoryRepository(client *db.InMemoryClient) Repository {
	return &InMemoryRepository{
		client: client,
	}
}
