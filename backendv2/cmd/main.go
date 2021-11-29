/*
  -name string
    	you name (default "eku@eku-HP-ProBook-450-G3")
  -port int
    	port that have to listen (default 35035)
*/
package main

import (
	"flag"
	"log"
	"os"
	"sync"

	config "github.com/aridae/p2p-messenger-coursework/backendv2/config"
	webserver "github.com/aridae/p2p-messenger-coursework/backendv2/ws-zefeer2peer-service"
)

var (
	peersPath *string
	portInt   *int
)

func init() {
	peersPath = flag.String("peers", "./backendv2/config/peers1.txt", "Path to file with peer addresses on each line")
	portInt = flag.Int("port", 35035, "port that have to listen")
	flag.Parse()

	// Настройки логирования
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	options := config.GetClientOptions()
	options.Port = *portInt
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
