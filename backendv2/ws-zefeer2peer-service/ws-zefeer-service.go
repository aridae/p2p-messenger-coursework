package wszefeer2peer

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"

	"github.com/aridae/p2p-messenger-coursework/backend/proto"
	config "github.com/aridae/p2p-messenger-coursework/backendv2/config"
	"github.com/aridae/p2p-messenger-coursework/backendv2/zefeer2peer"
	"github.com/gorilla/websocket"
	netp2p "github.com/libp2p/go-reuseport"
)

type BROWSER_COMMAND string

const (
	UNAME BROWSER_COMMAND = "UNAME" // send username to front
	PEERS BROWSER_COMMAND = "PEERS" // send peers to front
	MESSG BROWSER_COMMAND = "MESSG" // write message from front
)

type WSZefeerService struct {
	upgrader     websocket.Upgrader
	StaticServer *StaticServer
	ZefeerClient *zefeer2peer.ZefeerClient
}

func NewWSZefeerService(options *config.ClientOptions) *WSZefeerService {
	return &WSZefeerService{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		ZefeerClient: zefeer2peer.NewZefeerClient(options),
		StaticServer: NewStaticServer(),
	}
}

func (wszefeer *WSZefeerService) self() string {
	return "localhost:" + strconv.Itoa(wszefeer.ZefeerClient.Port)
}

func (wszefeer *WSZefeerService) InitServiceConnections(peersFile string) {
	file, err := os.Open(peersFile)
	if err != nil {
		log.Printf("Open peers.txt error: %s", err)
		return
	}

	var savedPeers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		savedPeers = append(savedPeers, scanner.Text())
	}

	self := "localhost:" + strconv.Itoa(wszefeer.ZefeerClient.Port)
	for _, peerAddress := range savedPeers {
		log.Printf("try to connect peer: %s", peerAddress)
		conn, err := netp2p.Dial("tcp", self, peerAddress)
		if err != nil {
			log.Printf("Dial ERROR: " + err.Error())
			return
		}
		newPeer := zefeer2peer.NewPeer(conn)
		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)
		readWriter := bufio.NewReadWriter(reader, writer)

		wszefeer.ZefeerClient.SendZPINGReq(newPeer)
		go wszefeer.ZefeerClient.HandleIncomingTraffic(readWriter, newPeer)
	}
}

func (wszefeer *WSZefeerService) CloseServiceConnections() {
	wszefeer.ZefeerClient.Peers.Empty()
}

func (wszefeer *WSZefeerService) StartServing(peersFile string) {

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		wszefeer.CloseServiceConnections()
		log.Printf("Exit by signal: %s", sig)
		os.Exit(1)
	}()

	log.Println("in StartServing")
	self := wszefeer.self()
	listener, err := netp2p.Listen("tcp", self)
	if err != nil {
		log.Printf("ListenTCP: %s", err.Error())
		os.Exit(1)
	}
	fmt.Printf("\n\tService start on %s\n\n", self)

	wszefeer.InitServiceConnections(peersFile)
	for {
		log.Println("accepting connections...")
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go wszefeer.onConnection(conn, wszefeer.ZefeerClient)
	}
}

func (wszefeer *WSZefeerService) HandleZefeerTraffic(rw *bufio.ReadWriter, conn net.Conn) {
	log.Println("in HandleZefeerTraffic")
	peer := zefeer2peer.NewPeer(conn)
	wszefeer.ZefeerClient.HandleIncomingTraffic(rw, peer)
}

func (wszefeer *WSZefeerService) HandleHTTPTraffic(rw *bufio.ReadWriter, conn net.Conn) {
	// считываем из ридера реквест
	request, err := http.ReadRequest(rw.Reader)
	if err != nil {
		log.Printf("Read request ERROR: %s", err)
		return
	}

	// подготовим респонс
	response := http.Response{
		StatusCode: 200,
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	// получили запрос от фронта на действия от зефир-клиента
	// для этого нужно смапить пришедший нам хттп запрос
	// на команды, которые известны нашему зефир-клиенту
	if path.Clean(request.URL.Path) == "/ws" {
		wszefeer.mapHTTPToZefeerTraffic(rw, NewWSZefeerWriter(conn), request)
		return
	}

	// получили запрос от фронта на статические данные
	wszefeer.StaticServer.ProcessStaticRequest(request, &response)
	err = response.Write(rw)
	if err != nil {
		log.Printf("Write response ERROR: %s", err)
		return
	}
	err = rw.Writer.Flush()
	if err != nil {
		log.Printf("Flush response ERROR: %s", err)
		return
	}
}

func (wszefeer *WSZefeerService) mapHTTPToZefeerTraffic(rw *bufio.ReadWriter, w http.ResponseWriter, r *http.Request) {
	wsconnection, err := wszefeer.upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer wsconnection.Close()
	defer log.Println("closing connection...")

	quit := make(chan bool)
	go wszefeer.waitMessageForWs(wsconnection, quit)
	for {
		wsMessageType, browserEnvelopeBytes, err := wsconnection.ReadMessage()
		if err != nil {
			log.Printf("ws read error: %v", err)
			break
		}
		log.Printf("got from brower: %s", browserEnvelopeBytes)

		browserEnvelope := &WSBrowserEnvelope{}
		err = json.Unmarshal(browserEnvelopeBytes, browserEnvelope)
		if err != nil {
			log.Printf("error on unmarshal message: %v", err)
			continue
		}

		// от фронта может прийти только реквест
		// на респонсы проверять не надо
		// ответы пишем в ws сокет браузера
		switch browserEnvelope.Cmd {
		// получили от браузера сообщение из поля ввода
		// надо отправить его пиру
		case string(MESSG):
			{
				fmt.Printf("JOPPA ENVELOPPA %+v\n", browserEnvelope)
				hexPubKey := browserEnvelope.To
				// if err != nil {
				// 	log.Printf("decode error: %s", err)
				// 	continue
				// }
				peer, found := wszefeer.ZefeerClient.Peers.Get(zefeer2peer.HashKey(hexPubKey))
				if found {
					fmt.Printf("JOPA FOUND!")
					wszefeer.ZefeerClient.SendMESSG(peer, browserEnvelope.Content)
					//wszefeer.ZefeerClient.HandleIncomingTraffic(rw, peer)
				}

				err := wsconnection.WriteMessage(wsMessageType, browserEnvelope.ToJson())
				if err != nil {
					log.Printf("ws write error: %s", err)
				}
			}
		// браузер запрашивает список пиров
		case string(PEERS):
			{
				log.Println("GOT PEERS FROM BROWSER")
				// TODO -
				// Request peers from our peers
				// Connect to these peers
				// Send peers to browser

				peers := wszefeer.ZefeerClient.Peers.ToList()
				log.Printf("our PEERS: %+v", peers)
				for _, p := range peers {
					fmt.Println("JO Pa !", p.PubKey)
				}

				wsPeerList := &WSPeerList{
					WSZefeerCmd: WSZefeerCmd{
						Cmd: string(PEERS),
					},
					Peers: peers,
				}
				peerListJson, _ := json.Marshal(wsPeerList)
				log.Printf("writing peers to wsconnection...")
				err := wsconnection.WriteMessage(wsMessageType, peerListJson)
				if err != nil {
					log.Printf("ws write error: %s", err)
				}
			}
		// браузер запрашиваем имя пользователя
		case string(UNAME):
			{
				unameEnvelope := WSUsername{
					WSZefeerCmd: WSZefeerCmd{
						string(UNAME),
					},
					Name:   wszefeer.ZefeerClient.Name,
					PubKey: string(wszefeer.ZefeerClient.PubKey),
				}
				err := wsconnection.WriteMessage(wsMessageType, unameEnvelope.ToJson())
				if err != nil {
					log.Printf("ws write error: %s", err)
				}
			}
		}
	}

	quit <- true
}

func (wszefeer *WSZefeerService) waitMessageForWs(wsconn *websocket.Conn, quit chan bool) {
	for {
		select {
		case envelope := <-wszefeer.ZefeerClient.Buffer:
			{
				log.Printf("New message: %s", envelope.Cmd)
				if string(envelope.Cmd) == string(zefeer2peer.MESSG) {
					wsCmd := proto.WsMessage{
						WsCmd: proto.WsCmd{
							Cmd: string(zefeer2peer.MESSG),
						},
						From:    hex.EncodeToString(envelope.From),
						To:      hex.EncodeToString(envelope.To),
						Content: string(envelope.Body),
					}

					wsCmdBytes, err := json.Marshal(wsCmd)
					if err != nil {
						panic(err)
					}
					err = wsconn.WriteMessage(1, wsCmdBytes)
					if err != nil {
						log.Printf("ws write error: %s", err)
					}
				}
			}
		case <-quit:
			{
				log.Printf("ws is broken")
				return
			}
		}
	}
}

// каждое новое соединение обрабатываем в отдельном потоке
func (wszefeer *WSZefeerService) onConnection(conn net.Conn, zefeer *zefeer2peer.ZefeerClient) {
	log.Printf("New connection from: %v", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	readWriter := bufio.NewReadWriter(reader, writer)

	// нужно получить первый четыре байта, чтобы проверить, является ли запрос
	// http запросом или зефир-запросом
	// http запрос приходит от фронта
	// зефир-запрос приходит от других пиров в сети
	first4bytes, err := readWriter.Peek(4)
	if err != nil {
		if err != io.EOF {
			log.Printf("Read peak ERROR: %s", err)
		}
		return
	}

	// если пришел запрос на соединение по http - значит, это от фронта:
	// фронт либо запросил статику, либо запросил действия от зефир-клиента
	if wszefeer.isHTTP(first4bytes) {
		wszefeer.HandleHTTPTraffic(readWriter, conn)
	} else {
		log.Println("HandleZefeerTraffic...")
		wszefeer.HandleZefeerTraffic(readWriter, conn)
	}
}

// проверяем, не прилетел ли нам хттп запрос от фронта
func (wszefeer *WSZefeerService) isHTTP(first4bytes []byte) bool {
	return map[string]bool{
		"GET ": true,
		"HEAD": true,
		"POST": true,
		"PUT ": true,
		"DELE": true,
		"CONN": true,
		"OPTI": true,
		"TRAC": true,
		"PATC": true,
	}[string(first4bytes)]
}
