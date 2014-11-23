package main

import (
    "github.com/googollee/go-socket.io"
    "log"
    "time"
)

type Room struct {
    Name string
    MessageChan chan string
    Range float64
}

// Returns a unique name for this group
// (and assumes the group names are unique, so it just returns the name)
func (r *Room) genUniqueName() (string) {
    return (*r).Name
}

func (r *Room) Run() {
    for {
        select {
            case message := <- (*r).MessageChan:
                log.Println("Group", (*r).Name, "received message", message)
            case <- time.After(5 * time.Minute):
                log.Println("Group", (*r).Name, "has been inactive for 5 minutes. Should we delete it?")
        }
    }
}

func (r Room) Join(so socketio.Socket) () {
    log.Println("A user has just joined group", r.Name)
}