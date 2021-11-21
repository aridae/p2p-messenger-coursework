package proto

import (
	"log"
	"time"

	"github.com/aridae/p2p-messenger-coursework/backendv2/domain"
)

func (p Proto) onHand(peer *domain.Peer, envelope *Envelope) {
	log.Printf("onHand")
	newPeer := domain.NewPeer(*peer.Conn)

	err := UpdatePeer(newPeer, envelope)
	if err != nil {
		log.Printf("Update peer error: %s", err)
	} else {
		if peer != nil {
			p.UnregisterPeer(peer)
		}

		peer.Name = newPeer.Name
		peer.PubKey = newPeer.PubKey
		peer.SharedKey = newPeer.SharedKey
		peer.LastSeen = time.Now().String()

		p.RegisterPeer(peer)
	}
	p.SendName(peer)
}

func (p Proto) onMess(peer *domain.Peer, envelope *Envelope) {
	envelope.Content = Decrypt(envelope.Content, peer.SharedKey.Secret)
	p.Broker <- envelope
}

func (p Proto) onList(peer *domain.Peer, envelope *Envelope) {
	log.Printf("onList")
}
