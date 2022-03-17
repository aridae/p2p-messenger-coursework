package wszefeer2peer

import (
	"encoding/json"

	"github.com/aridae/p2p-messenger-coursework/backend/v2/zefeer2peer"
)

// тут модели, которые могут прийти от бразуера
type Serializable interface {
	ToJson() []byte
}

func toJson(v interface{}) []byte {
	json, err := json.Marshal(v)

	if err != nil {
		panic(err)
	}

	return json
}

//WSZefeerCmd WebSocket command
type WSZefeerCmd struct {
	Cmd string `json:"cmd"`
}

//WSUsername WebSocket command: PeerName
type WSUsername struct {
	WSZefeerCmd
	Name   string `json:"name"`
	PubKey string `json:"id"`
}

//ToJson convert to JSON bytes
func (v WSUsername) ToJson() []byte {
	return toJson(v)
}

//WSPeerList WebSocket command: list of peers
type WSPeerList struct {
	WSZefeerCmd
	Peers []zefeer2peer.PeerName `json:"peers"`
}

//ToJson convert to JSON bytes
func (v WSPeerList) ToJson() []byte {
	return toJson(v)
}

//WSBrowserEnvelope WebSocket command: new Message
type WSBrowserEnvelope struct {
	WSZefeerCmd
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
}

//ToJson convert to JSON bytes
func (v WSBrowserEnvelope) ToJson() []byte {
	return toJson(v)
}
