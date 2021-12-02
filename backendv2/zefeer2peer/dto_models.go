package zefeer2peer

import "encoding/json"

//Serializable interface to detect that can to serialised to json
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

//PeerName Peer name and public key
type PeerName struct {
	Name   string `json:"name"`
	PubKey string `json:"id"`
}

//ToJson convert to JSON bytes
func (v PeerName) ToJson() []byte {
	return toJson(v)
}

type PeerZPING struct {
	Name   string `json:"name"`
	PubKey string `json:"id"`
	ExKey  string `json:"exKey"`
}

//ToJson convert to JSON bytes
func (v PeerZPING) ToJson() []byte {
	return toJson(v)
}

type PeerMESSG struct {
	Name    string `json:"name"`
	PubKey  string `json:"id"`
	ExKey   string `json:"exKey"`
	Message string `json:"message"`
}

type PeerMESSGBuffed struct {
	Name    string `json:"name"`
	PubKey  string `json:"id"`
	ExKey   string `json:"exKey"`
	Message string `json:"message"`
}

//ToJson convert to JSON bytes
func (v PeerMESSG) ToJson() []byte {
	return toJson(v)
}

type PeerPEERSReq struct {
	Name   string `json:"name"`
	PubKey string `json:"id"`
	ExKey  string `json:"exKey"`
}

//ToJson convert to JSON bytes
func (v PeerPEERSReq) ToJson() []byte {
	return toJson(v)
}

type PeerPEERSResp struct {
	Name   string `json:"name"`
	PubKey string `json:"id"`
	ExKey  string `json:"exKey"`

	PeerNames []PeerName `json:"peers"`
}

//ToJson convert to JSON bytes
func (v PeerPEERSResp) ToJson() []byte {
	return toJson(v)
}
