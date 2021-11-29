package zefeer2peer

import (
	"bufio"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"time"

	config "github.com/aridae/p2p-messenger-coursework/backendv2/config"
)

type COMMAND string

const (
	ZPING COMMAND = "ZPING" // send name, accept name
	PEERS COMMAND = "PEERS" // query peers of a particular peer
	MESSG COMMAND = "MESSG" // send message
)

type COMMAND_TYPE string

const (
	REQUEST  COMMAND_TYPE = "REQS"
	RESPONSE COMMAND_TYPE = "RESP"
)

type ZefeerClient struct {
	Port    int
	Host    string
	Name    string
	PubKey  ed25519.PublicKey
	privKey ed25519.PrivateKey

	// WIP
	Buffer    chan *Envelope
	BufferOut chan *Envelope

	// все пиры vs пиров, с которыми текущий коннект
	Peers *PeersHashTable
}

func NewZefeerClient(options *config.ClientOptions) *ZefeerClient {
	publicKey, privateKey := LoadKey(options.Username)
	client := &ZefeerClient{
		Port:    options.Port,
		Name:    options.Username,
		Peers:   NewPeersHashTable(),
		PubKey:  publicKey,
		privKey: privateKey,
		Buffer:  make(chan *Envelope),
	}

	return client
}

func (p ZefeerClient) RegisterPeer(peer *Peer) *Peer {
	log.Println("in RegisterPeer")
	if reflect.DeepEqual(peer.PubKey, p.PubKey) {
		return nil
	}

	p.Peers.Put(peer)
	log.Printf("Register new peer: %s (%v)", peer.Name, len(p.Peers.peers))
	return peer
}

func (p ZefeerClient) UnregisterPeer(peer *Peer) {
	log.Println("in UnregisterPeer")
	if p.Peers.Remove(peer) {
		log.Printf("UnRegister peer: %s", peer.Name)
	}
}

// два варианта возможны в обработке входящяего и исходящего траффика
// 1 - один тип запроса для входящего и исходящего траффика
//   - преимущества: единообразие
//   - недостатки: если нам пришел пакет, надо как-то определить, это ответ на наш запрос или
//   самостоятельный запрос нам, потому что сокет у нас один и для входящего, и для исходящего траффика
//   причем пакеты разным клиентам приходят и отправляются асинхронно в разных потоках
// 2 - под каждый тип запроса подтип для входящего и исходящего траффика
// причем для каждого конверта может быть сгенерирован идентификатор, для
// восстановления соответствия запрос-ответ - выбран этот вариант

// сокет может принимать соединения от многих пиров
// потому что создается копия сокета, которая остается в состоянии прослушивания
// для принятия новых подключений
// получается можно:
// 1 - принять или инициализировать подключение к сокету
// 2 - выделить отдельный поток для обработки копии сокета
// 3 - там обрабатывать это соединение, пока оно не будет закрыто
//     нами или другим пиром
func (p ZefeerClient) SendZPINGReq(peer *Peer) {
	log.Println("in SendZPINGReq")
	exchPubKey, exchPrivKey := CreateKeyExchangePair()
	zping := PeerZPING{
		Name:   p.Name,
		PubKey: hex.EncodeToString(p.PubKey),
		ExKey:  hex.EncodeToString(exchPubKey[:]),
	}.ToJson()

	peer.SharedKey.Update(nil, exchPrivKey[:])
	sign := ed25519.Sign(p.privKey, zping)
	envelope := NewSignedEnvelope(string(ZPING), string(REQUEST), p.PubKey[:], make([]byte, 32), sign, zping)
	envelope.Send(peer)
}

func (p ZefeerClient) SendZPINGResp(peer *Peer) {
	log.Println("in SendZPINGResp")
	exchPubKey, exchPrivKey := CreateKeyExchangePair()
	zping := PeerZPING{
		Name:   p.Name,
		PubKey: hex.EncodeToString(p.PubKey),
		ExKey:  hex.EncodeToString(exchPubKey[:]),
	}.ToJson()

	peer.SharedKey.Update(nil, exchPrivKey[:])
	sign := ed25519.Sign(p.privKey, zping)
	envelope := NewSignedEnvelope(string(ZPING), string(RESPONSE), p.PubKey[:], make([]byte, 32), sign, zping)
	envelope.Send(peer)
}

func (p ZefeerClient) SendPEERSReq(peer *Peer) {
	exchPubKey, exchPrivKey := CreateKeyExchangePair()
	peersReq := PeerPEERSReq{
		Name:   p.Name,
		PubKey: hex.EncodeToString(p.PubKey),
		ExKey:  hex.EncodeToString(exchPubKey[:]),
	}.ToJson()

	peer.SharedKey.Update(nil, exchPrivKey[:])
	sign := ed25519.Sign(p.privKey, peersReq)
	envelope := NewSignedEnvelope(string(PEERS), string(REQUEST), p.PubKey[:], make([]byte, 32), sign, peersReq)
	envelope.Send(peer)
}

func (p ZefeerClient) SendPEERSResp(peer *Peer) {
	exchPubKey, exchPrivKey := CreateKeyExchangePair()
	peersResp := PeerPEERSResp{
		Name:      p.Name,
		PubKey:    hex.EncodeToString(p.PubKey),
		ExKey:     hex.EncodeToString(exchPubKey[:]),
		PeerNames: p.Peers.ToList(),
	}.ToJson()

	peer.SharedKey.Update(nil, exchPrivKey[:])
	sign := ed25519.Sign(p.privKey, peersResp)
	envelope := NewSignedEnvelope(string(PEERS), string(REQUEST), p.PubKey[:], make([]byte, 32), sign, peersResp)
	envelope.Send(peer)
}

func (p ZefeerClient) SendMESSG(peer *Peer, message string) {
	exchPubKey, exchPrivKey := CreateKeyExchangePair()
	peersMesg := PeerMESSG{
		Name:    p.Name,
		PubKey:  hex.EncodeToString(p.PubKey),
		ExKey:   hex.EncodeToString(exchPubKey[:]),
		Message: message,
	}.ToJson()

	peer.SharedKey.Update(nil, exchPrivKey[:])
	sign := ed25519.Sign(p.privKey, peersMesg)

	envelope := NewSignedEnvelope(string(PEERS), string(REQUEST), p.PubKey[:], make([]byte, 32), sign, peersMesg)
	envelope.Send(peer)
}

// эта часть относится к пиру как к серверу
// мы получили конверт envelope c  командой з-пинг
// - проверяем подпись конверта
// - получаем имя пира, который его отправил
// - обновляем информацию о нем в базе пиров
// - проверяем, есть ли он в наших пирах, возможно добавляем
// - отправляем конверт с нашим именем и меткой з-пинг в ответ
func (zefeer ZefeerClient) onZPINGReq(peer *Peer, envelope *Envelope) {
	log.Println("onZPINGReq")

	// получили запрос на зпинг от другого пира
	newPeer := NewPeer(*peer.Conn)
	err := newPeer.UpdatePeerOnZPING(envelope)
	if err != nil {
		log.Printf("Update peer error: %s", err)
	} else {
		oldPeer, found := zefeer.Peers.Get(HashKey(peer.PubKey))
		if found {
			oldPeer.Name = newPeer.Name
			oldPeer.PubKey = newPeer.PubKey
			oldPeer.SharedKey = newPeer.SharedKey
			oldPeer.LastSeen = time.Now().String()
		} else {
			zefeer.RegisterPeer(peer)
		}
	}

	// отправляем ему ответ
	zefeer.SendZPINGResp(peer)
}

func (zefeer ZefeerClient) onZPINGResp(peer *Peer, envelope *Envelope) {
	log.Println("onZPINGResp")

	// получили ответ на наш зпинг
	// обновляем информацию о пирах и все
	newPeer := NewPeer(*peer.Conn)
	err := newPeer.UpdatePeerOnZPING(envelope)
	if err != nil {
		log.Printf("Update peer error: %s", err)
	} else {
		oldPeer, found := zefeer.Peers.Get(HashKey(peer.PubKey))
		if found {
			oldPeer.Name = newPeer.Name
			oldPeer.PubKey = newPeer.PubKey
			oldPeer.SharedKey = newPeer.SharedKey
			oldPeer.LastSeen = time.Now().String()
		} else {
			zefeer.RegisterPeer(peer)
		}
	}
}

// мы получили конверт envelope c  командой peers
// - проверяем подпись конверта
// - берем имя пира
// - отправляем ему наших пиров
func (zefeer ZefeerClient) onPEERSReq(peer *Peer, envelope *Envelope) {
	log.Println("onPEERSReq")
	zefeer.SendPEERSResp(peer)
}

func (zefeer ZefeerClient) onPEERSResp(peer *Peer, envelope *Envelope) {
	log.Printf("onPEERSResp: %s\n", string(envelope.Body))
	envelope.Body = Decrypt(envelope.Body, peer.SharedKey.Secret)

	// envelope body to json
	var peersResp PeerPEERSResp
	if err := json.Unmarshal(envelope.Body, &peersResp); err != nil {
		log.Println("unmarshalling error")
		return
	}
	fmt.Printf("Got peers response %+v\n", peersResp)
}

// мы получили конверт envelope c  командой MESSG
// - проверяем подпись конверта
// - берем имя пира
// - обновляем информацию о нем в базе пиров
// - читаем сообщение в буфер

// по принятию сообщения, надо либо пушить на фронт сразу
// либо в буфер(!) - по открытию окна во фронте, фронт может
// обращаться к этому буферы
func (zefeer ZefeerClient) onMESSG(peer *Peer, envelope *Envelope) {
	log.Println("onMESSG")
	envelope.Body = Decrypt(envelope.Body, peer.SharedKey.Secret)

	log.Println("Writing message to buffer...")
	var peersMsg PeerMESSG
	if err := json.Unmarshal(envelope.Body, &peersMsg); err != nil {
		log.Println("unmarshalling error")
		return
	}
	fmt.Printf("Got message %+v\n", peersMsg)
	zefeer.Buffer <- envelope
}

func (p ZefeerClient) HandleIncomingTraffic(rw *bufio.ReadWriter, peer *Peer) {
	log.Println("in HandleIncomingTraffic")

	// обрабатываем траффик, пока кто-то не прервет соединение
	for {
		envelope, err := ReadEnvelope(rw.Reader)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error on read Envelope: %v", err)
			}
			log.Printf("Disconnect peer %s due to connection break", peer)
			break
		}
		if ed25519.Verify(envelope.From, envelope.Body, envelope.Sign) {
			log.Printf("Signed envelope!")
		}

		log.Printf("LISTENER: receive envelope [%+v] from %s", envelope, (*peer.Conn).RemoteAddr())
		switch string(envelope.Cmd) {
		case string(ZPING):
			log.Printf("LISTENER: got ZPING from %s", (*peer.Conn).RemoteAddr())
			if string(envelope.CmdType) == string(REQUEST) {
				p.onZPINGReq(peer, envelope)
			} else {
				p.onZPINGResp(peer, envelope)
			}
		case string(PEERS):
			log.Printf("LISTENER: got PEERS from %s", (*peer.Conn).RemoteAddr())
			if string(envelope.CmdType) == string(REQUEST) {
				p.onPEERSReq(peer, envelope)
			} else {
				p.onPEERSResp(peer, envelope)
			}
		case string(MESSG):
			log.Printf("LISTENER: got MESSG from %s", (*peer.Conn).RemoteAddr())
			p.onMESSG(peer, envelope)
		default:
			log.Printf("LISTENER: unknown command %s", (*peer.Conn).RemoteAddr())
		}
	}
}

func (p ZefeerClient) ListenPeer(peer *Peer) {
	log.Println("in ListenPeer")

	readWriter := bufio.NewReadWriter(bufio.NewReader(*peer.Conn), bufio.NewWriter(*peer.Conn))
	p.HandleIncomingTraffic(readWriter, peer)
}
