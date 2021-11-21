package peerrepository

import (
	"github.com/aridae/p2p-messenger-coursework/backendv2/db"
	"github.com/aridae/p2p-messenger-coursework/backendv2/domain"
)

type SQLiteRepository struct {
	client *db.SQLiteClient
}

func (r *SQLiteRepository) ListPeers() ([]domain.Peer, error) {
	return nil, nil
}

func (r *SQLiteRepository) GetPeer(pubKey string) (*domain.Peer, error) {
	return nil, nil
}

func (r *SQLiteRepository) RemovePeer(pubKey string) error {
	return nil
}

func (r *SQLiteRepository) AddPeer(*domain.Peer) (string, error) {
	return "testKey", nil
}

func NewSQLiteRepository(client *db.SQLiteClient) Repository {
	return &SQLiteRepository{
		client: client,
	}
}
