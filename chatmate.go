package main

import (
	"encoding/json"
	"github.com/googollee/go-socket.io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Should look like path
const websocketRoom = "/chat"

type initMsg struct {
	username  string
	latitude  float64
	longitude float64
}

func main() {
	lastMessages := []string{}
	var lmMutex sync.Mutex

	// Configuring socket.io Server
	sio, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	sio.On("connection", func(so socketio.Socket) {
		var username string
		username = "User-" + so.Id()
		log.Println("on connection", username)
		so.Join(websocketRoom)

		lmMutex.Lock()
		for i, _ := range lastMessages {
			so.Emit("message", lastMessages[i])
		}
		lmMutex.Unlock()

		so.On("joined_message", func(message string) {
			log.Println("w000t")
			log.Println(message)

			decoder := json.NewDecoder(strings.NewReader(message))
			var msg initMsg
			decoder.Decode(&msg)

			log.Printf("joined_message %+v\n", msg)
			res := map[string]interface{}{
				"username":  msg.username,
				"dateTime":  time.Now().UTC().Format(time.RFC3339Nano),
				"latitude":  msg.latitude,
				"longitude": msg.longitude,
				"type":      "joined_message",
			}
			jsonRes, _ := json.Marshal(res)
			so.Emit("message", string(jsonRes))
			so.BroadcastTo(websocketRoom, "message", string(jsonRes))
		})
		so.On("send_message", func(message string) {
			log.Println("send_message from", username)
			res := map[string]interface{}{
				"username": username,
				"message":  message,
				"dateTime": time.Now().UTC().Format(time.RFC3339),
				"type":     "message",
			}
			jsonRes, _ := json.Marshal(res)
			lmMutex.Lock()
			if len(lastMessages) == 100 {
				lastMessages = lastMessages[1:100]
			}
			lastMessages = append(lastMessages, string(jsonRes))
			lmMutex.Unlock()
			so.Emit("message", string(jsonRes))
			so.BroadcastTo(websocketRoom, "message", string(jsonRes))
		})
		so.On("disconnection", func() {
			log.Println("on disconnect", username)
		})
	})
	sio.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	// Sets up the handlers and listen on port 8080
	http.Handle("/socket.io/", sio)
	// Serve CSS, Javascript...
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))
	// Serve index.html
	http.Handle("/", http.FileServer(http.Dir("./public/")))

	port := ":1337"
	log.Println("Server started on port", port)

	// "The handler is usually nil, which means to use DefaultServeMux." - golang.org
	log.Fatal(http.ListenAndServe(port, nil))
}
