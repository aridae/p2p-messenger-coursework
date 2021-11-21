package domain

type Message struct {
	From    *Peer
	To      *Peer
	Content string
}
