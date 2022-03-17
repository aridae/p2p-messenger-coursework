package zefeer2peer

import (
	"strings"
	"time"
)

type MessageBuffed struct {
	Datetime string
	IsSelf   bool
	Content  string
}

func (v MessageBuffed) ToJson() []byte {
	return toJson(v)
}

func (client *ZefeerClient) MessageFromEnvelope(envelope *Envelope) *MessageBuffed {
	return &MessageBuffed{
		Datetime: time.Now().Format("dd-MM-yyyy HH:mm"),
		IsSelf:   strings.EqualFold(string(envelope.From), string(client.PubKey)),
		Content:  string(envelope.Body),
	}
}
