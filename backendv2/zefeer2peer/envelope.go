package zefeer2peer

// поскольку у нас протокол для обмена сообщениями,
// то по аналогии с конвертами для SMTP была предложена
// идея использовать конверты для сообщений нашего zefeer2peer

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"log"
	"strconv"
)

// размер хэдера фиксированный,
// размер боди указывается в самом конверте
var cmdLen = 5
var ctypeLen = 4
var idLen = 16
var fromLen = 32
var toLen = 32
var signLen = 64
var headerLen = cmdLen + ctypeLen + idLen + fromLen + toLen + signLen + 2

type Header struct {
	Cmd     []byte // PEERS, ZEFIR, MESSG
	CmdType []byte // REQS, RESP
	Id      []byte
	From    []byte
	To      []byte
	Sign    []byte
}

//Envelope Конверт для сообщений между пирами
// Cmd - команда - PEERS, ZEFIR, MESSG
type Envelope struct {
	Header
	BodyLength uint16
	Body       []byte
}

func (v Envelope) ToJson() []byte {
	return toJson(v)
}

func (m Envelope) String() string {
	return string(m.Cmd) + "-" + string(m.Id) + "-" + strconv.FormatUint(uint64(m.BodyLength), 10)
}

func getRandomSeed(l int) []byte {
	seed := make([]byte, l)
	_, err := rand.Read(seed)
	if err != nil {
		log.Printf("rand.Read Error: %v", err)
	}
	return seed
}

//NewEnvelope Создание нового конверта
func NewEnvelope(cmd string, cmd_type string, contentBytes []byte) (envelope *Envelope) {
	contentLength := len(contentBytes)
	if contentLength >= 65535 {
		contentBytes = contentBytes[:65535]
	}

	envelope = &Envelope{
		Header: Header{
			Cmd:     []byte(cmd)[:cmdLen],
			CmdType: []byte(cmd_type)[:ctypeLen],
			Id:      getRandomSeed(idLen)[:idLen],
			From:    make([]byte, fromLen),
			To:      make([]byte, toLen),
			Sign:    make([]byte, signLen),
		},
		BodyLength: uint16(contentLength),
		Body:       contentBytes[0:contentLength],
	}
	return
}

//NewSignedEnvelope create new envelope with signature
func NewSignedEnvelope(cmd string, cmd_type string, from []byte, to []byte, sign []byte, contentBytes []byte) (envelope *Envelope) {
	envelope = NewEnvelope(cmd, cmd_type, contentBytes)
	envelope.From = from
	envelope.To = to
	envelope.Sign = sign
	return
}

func (m Envelope) Serialize() []byte {
	result := make([]byte, 0, headerLen+len(m.Body))
	result = append(result, m.Cmd[0:cmdLen]...)
	result = append(result, m.CmdType[0:ctypeLen]...)
	result = append(result, m.Id[0:idLen]...)
	result = append(result, m.From[0:fromLen]...)
	result = append(result, m.To[0:toLen]...)
	result = append(result, m.Sign[0:signLen]...)

	contentLengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(contentLengthBytes, m.BodyLength)

	result = append(result, contentLengthBytes...)
	result = append(result, m.Body...)
	return result
}

//UnSerialize Десериализация массива байт в конверт с содержимым
func UnSerialize(b []byte) (envelope *Envelope) {
	//log.Printf("in UnSerialize: %d, %s\n", len(b), b[cmdLen+ctypeLen+idLen+fromLen+toLen:])
	contentLength := binary.BigEndian.Uint16(b[headerLen-2 : headerLen])
	if contentLength > 65535 {
		return nil
	}

	envelope = &Envelope{
		Header: Header{
			Cmd:     b[0:cmdLen],
			CmdType: b[cmdLen : cmdLen+ctypeLen],
			Id:      b[cmdLen+ctypeLen : cmdLen+ctypeLen+idLen],
			From:    b[cmdLen+ctypeLen+idLen : cmdLen+ctypeLen+idLen+fromLen],
			To:      b[cmdLen+ctypeLen+idLen+fromLen : cmdLen+ctypeLen+idLen+fromLen+toLen],
			Sign:    b[cmdLen+ctypeLen+idLen+fromLen+toLen : cmdLen+ctypeLen+idLen+fromLen+toLen+signLen],
		},
		BodyLength: contentLength,
	}
	if len(b) == (headerLen + int(contentLength)) {
		envelope.Body = b[headerLen:]
	} else {
		envelope.Body = make([]byte, contentLength)
	}
	return
}

//ReadEnvelope Формирование конверта из байтов ридера сокета
func ReadEnvelope(reader *bufio.Reader) (*Envelope, error) {
	log.Println("in ReadEnvelope")
	header := make([]byte, headerLen)

	// read envelope header
	_, err := reader.Read(header)
	if err != nil {
		return nil, err
	}

	envelope := UnSerialize(header)
	_, err = reader.Read(envelope.Body)
	if err != nil {
		return nil, err
	}

	return envelope, nil
}

//Send send envelope to peer
func (m Envelope) Send(peer *Peer) {
	log.Printf("Send %s [%s] to peer %s ", m.Cmd, m.CmdType, peer.Name)
	_, err := (*peer.Conn).Write(m.Serialize())
	if err != nil {
		log.Printf("ERROR on write message: %v", err)
	}
}
