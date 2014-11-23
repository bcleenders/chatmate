package main

import (
	"encoding/json"
	"github.com/googollee/go-socket.io"
    // "github.com/kellydunn/golang-geo"
	"log"
	"net/http"
    "sync"
    "strings"
    "time"
    // "main"
)

// Should look like path
const websocketRoom = "/chat"

type initMsg struct {
	Username  string
	Latitude  float64
	Longitude float64
}

var rooms map[string]Room = make(map[string]Room)

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
			decoder := json.NewDecoder(strings.NewReader(message))
			var msg initMsg
			decoder.Decode(&msg)

			log.Printf("joined_message %+v\n", msg)
			res := map[string]interface{}{
				"username":  msg.Username,
				"dateTime":  time.Now().UTC().Format(time.RFC3339Nano),
				"latitude":  msg.Latitude,
				"longitude": msg.Longitude,
				"type":      "joined_message",
			}
			jsonRes, _ := json.Marshal(res)
			so.Emit("message", string(jsonRes))
			so.BroadcastTo(websocketRoom, "message", string(jsonRes))

            for k, v := range rooms {
                if (true) { // Should be a check if the group is in range!
                    info := map[string]interface{}{
                        "name": k,
                        "range": v.Range,
                        "people": 40,
                        "type": "group_info",
                    }
                    jsonInfo, _ := json.Marshal(info)
                    so.Emit("group_info", string(jsonInfo))
                }
            }
		})
		so.On("send_message", func(message string) {
			log.Println("send_message from", username)

            type Message struct {
                Group string
                Message string
            }

            var msg Message
            decoder := json.NewDecoder(strings.NewReader(message))
            decoder.Decode(&msg)

			res := map[string]interface{}{
				"username": username,
                "group": msg.Group,
				"message":  msg.Message,
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
            so.BroadcastTo(msg.group, "message", string(jsonRes))
			// so.BroadcastTo(websocketRoom, "message", string(jsonRes))
		})
        so.On("new_group", func(message string) {
            // Create the new group
            messageChan := make(chan string)
            var newRoom Room

            decoder := json.NewDecoder(strings.NewReader(message))
            decoder.Decode(&newRoom)
            newRoom.MessageChan = messageChan
            rooms[newRoom.Name] = newRoom
            rooms[newRoom.Name].Join(so) 

            go newRoom.Run()
        })
        so.On("join_group", func(message string) {
            so.Join(message)
            rooms[message].Join(so)
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
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("/home/damnyankee/public/"))))
	// Serve index.html
	http.Handle("/", http.FileServer(http.Dir("/home/damnyankee/public/")))

	port := ":1337"
	log.Println("Server started on port", port)

	// "The handler is usually nil, which means to use DefaultServeMux." - golang.org
	log.Fatal(http.ListenAndServe(port, nil))
}
