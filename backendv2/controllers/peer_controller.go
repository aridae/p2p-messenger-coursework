package controllers

import (
	peerrepo "github.com/aridae/p2p-messenger-coursework/backendv2/data_access/peer_repository"
	"github.com/aridae/p2p-messenger-coursework/backendv2/domain"
)

type PeerController struct {
	PeerRepo peerrepo.Repository
}

func (c PeerController) GetPeer(pubKey string) (*domain.Peer, error) {
	return c.PeerRepo.GetPeer(pubKey)
}

func (c PeerController) ListPeers() ([]domain.Peer, error) {
	return c.PeerRepo.ListPeers()
}

func (c PeerController) AddPeer(peer *domain.Peer) (string, error) {
	return c.PeerRepo.AddPeer(peer)
}

func (c PeerController) RemovePeer(pubKey string) error {
	return c.PeerRepo.RemovePeer(pubKey)
}
