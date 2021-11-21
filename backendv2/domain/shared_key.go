package domain

import (
	"github.com/aridae/p2p-messenger-coursework/backendv2/utils/cipher"
)

// TODO: это кто и зачем?
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
		secret := cipher.CalcSharedSecret(sk.RemoteKey, sk.LocalKey)
		sk.Secret = secret[:32]
	}
}
