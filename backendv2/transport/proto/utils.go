package proto

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/aridae/p2p-messenger-coursework/backendv2/domain"
)

//UpdatePeer Update peer struct after handshake
func UpdatePeer(p *domain.Peer, envelope *Envelope) error {
	if string(envelope.Cmd) != "HAND" {
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
