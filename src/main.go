package main

import (
	"flag"
	"go_test/pqueue"
	"go_test/services"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	log.Println("Messages handling started")
	start := time.Now()

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// define wss connection from
	url_from := url.URL{Scheme: "wss", Host: "test-ws.skns.dev", Path: "/raw-messages"}

	log.Printf("connecting to %s", url_from.String())

	from, _, err := websocket.DefaultDialer.Dial(url_from.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer from.Close()

	// define wss connection to
	url_to := url.URL{Scheme: "wss", Host: "test-ws.skns.dev", Path: "/raw-messages/naguslaev"}
	log.Printf("connecting to %s", url_to.String())

	to, _, err := websocket.DefaultDialer.Dial(url_to.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer to.Close()

	done := make(chan struct{})
	queue := &pqueue.PriorityQueueMutex{}

	sender := services.NewMessageSender(to, done, queue)
	loader := services.NewMessageLoader(from, done, queue, sender)

	loader.Load()
	sender.Send()

	for {
		select {
		case <-done:
			log.Println("Spended time:", time.Since(start))
			log.Println("Minimum delay: ", sender.GetMinLatency())
			log.Println("Messages sent: ", sender.GetSentCount())
			// fmt.Println("Messages instantly sent: ", loader.GetInstantCount())
			// fmt.Println("Messages queued: ", loader.GetQueuedCount())
			os.Exit(0)
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := from.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("wss FROM close:", err)
				return
			}
			err = to.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("wss TO close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
