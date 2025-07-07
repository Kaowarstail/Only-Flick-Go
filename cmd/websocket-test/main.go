package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var token = flag.String("token", "", "JWT token for authentication")

func main() {
	flag.Parse()
	log.SetFlags(0)

	if *token == "" {
		log.Fatal("JWT token is required. Use -token flag")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/api/v1/ws"}
	log.Printf("Connecting to %s", u.String())

	// Ajouter le token JWT dans les headers
	headers := map[string][]string{
		"Authorization": {"Bearer " + *token},
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Envoyer un message de test toutes les 10 secondes
	testMessage := `{
		"type": "heartbeat",
		"data": {
			"client_time": "%s"
		}
	}`

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			message := fmt.Sprintf(testMessage, time.Now().Format(time.RFC3339))
			err := c.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
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
