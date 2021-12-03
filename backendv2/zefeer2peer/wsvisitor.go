package zefeer2peer

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

const (
	ZefeerMessageType = websocket.TextMessage
	MESSGCMD          = "MESSG"
)

type WSVisitor struct {
	conn *websocket.Conn
}

type WSBrowserEnvelope struct {
	Cmd     string `json:"cmd"`
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
}

func (v WSBrowserEnvelope) ToJson() []byte {
	return toJson(v)
}

func NewWSVisitor() *WSVisitor {
	return &WSVisitor{}
}

func (visitor *WSVisitor) UpdateOnConnect(newConn *websocket.Conn) {
	log.Printf("OBNOVA VISITORA !")

	visitor.conn = newConn
}

func (visitor *WSVisitor) VisitOnMESSG(env *Envelope) {
	if visitor.conn != nil {
		log.Printf("VISITORA JOPU VIDNO!!!!!!")

		var peersMsg PeerMESSG
		if err := json.Unmarshal(env.Body, &peersMsg); err != nil {
			log.Println("unmarshalling error")
			return
		}

		wsEnvelope := WSBrowserEnvelope{
			Cmd:     MESSGCMD,
			From:    string(env.From),
			To:      string(env.To),
			Content: string(peersMsg.Message),
		}
		err := visitor.conn.WriteMessage(ZefeerMessageType, wsEnvelope.ToJson())
		if err != nil {
			log.Printf("ws write error: %s", err)
		}
	} else {
		log.Printf("VISITORA NE VIDNO!")
	}
}
