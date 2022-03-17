/*
Go training project

  -name string
    	you name (default "eku@eku-HP-ProBook-450-G3")
  -peers string
    	Path to file with peer addresses on each line (default "peers.txt")
  -port int
    	port that have to listen (default 35035)
  -webview
    	Start WebView ui (default false)
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"sync"
	"syscall"

	"github.com/aridae/p2p-messenger-coursework/backend/v1/discover"
	"github.com/aridae/p2p-messenger-coursework/backend/v1/listener"
	"github.com/aridae/p2p-messenger-coursework/backend/v1/proto"
	"github.com/webview/webview"
)

//InitParams initializing params
type InitParams struct {
	Name         *string
	Port         *int
	PeersFile    *string
	StartWebView *bool
}

var initParams InitParams

func init() {
	currentUser, _ := user.Current()
	hostName, _ := os.Hostname()

	initParams = InitParams{
		Name:         flag.String("name", currentUser.Username+"@"+hostName, "you name"),
		Port:         flag.Int("port", 35035, "port that have to listen"),
		PeersFile:    flag.String("peers", "./backend/config/peers.txt", "Path to file with peer addresses on each line"),
		StartWebView: flag.Bool("webview", false, "Start WebView ui"),
	}

	flag.Parse()

	//flag.PrintDefaults()

	// Настройки логирования
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

}

func main() {

	fff := retro()

	i := fff(1, 2)
	log.Printf("result = %v", i)

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		log.Printf("Exit by signal: %s", sig)
		os.Exit(1)
	}()

	p := proto.NewProto(*initParams.Name, *initParams.Port)

	if *initParams.StartWebView {
		startWithWebView(p)
	} else {
		startWithoutWebView(p)
	}
}

func retro() func(a int, b int) int {
	return func(a int, b int) int {
		return a + b
	}
}

//startWithWebView Запуск приложения с WebView
func startWithWebView(p *proto.Proto) {
	go discover.StartDiscover(p, *initParams.PeersFile)
	go listener.StartListener(p, *initParams.Port)

	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Peer To Peer Messenger")
	// w.SetSize(800, 600, webview.HintNone)
	w.Navigate(fmt.Sprintf("http://localhost:%v", initParams.Port))
	w.Run()

}

//startWithoutWebView Запуск приложения без запуска WebView
func startWithoutWebView(p *proto.Proto) {
	var wg sync.WaitGroup
	wg.Add(2)
	go discover.StartDiscover(p, *initParams.PeersFile)
	go listener.StartListener(p, *initParams.Port)
	wg.Wait()
}
