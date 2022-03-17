package main

import (
	"flag"
	"log"
	"os"
	"sync"

	config "github.com/aridae/p2p-messenger-coursework/backend/v2/config"
	webserver "github.com/aridae/p2p-messenger-coursework/backend/v2/ws-zefeer2peer-service"
)

var (
	peersPath *string
	portInt   *int
	uname     *string
)

func init() {
	uname = flag.String("name", "zefeerchik", "your name")
	peersPath = flag.String("peers", "./backendv2/config/peers1.txt", "Path to file with peer addresses on each line")
	portInt = flag.Int("port", 35035, "port that have to listen")
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	options := config.GetClientOptions()
	options.Port = *portInt
	options.Username = *uname
	wszefeer := webserver.NewWSZefeerService(options)
	startWithoutWebView(wszefeer)
}

//startWithoutWebView Запуск приложения без запуска WebView
func startWithoutWebView(zefeer *webserver.WSZefeerService) {
	var wg sync.WaitGroup
	wg.Add(2)
	//go zefeer2peer.StartLookup(zefeer.ZefeerClient, *peersPath)
	go zefeer.StartServing(*peersPath)
	wg.Wait()
}
