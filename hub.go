package main

import (
	// "fmt"
	"log"
	// "quickchat/database"
	"time"
)

// Hub websocket hub
type Hub struct {
	pass       string
	chatid     int
	users      map[*Profile]bool
	broadcast  chan []byte
	register   chan *Profile
	unregister chan *Profile
}

func newHub(id int, key string) *Hub {
	return &Hub{
		pass:       key,
		chatid:     id,
		users:      make(map[*Profile]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Profile),
		unregister: make(chan *Profile),
	}
}

func (hub *Hub) run() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case newUser := <-hub.register:
			hub.users[newUser] = true
			NumberOfConnections++

		case delUser := <-hub.unregister:
			if _, ok := hub.users[delUser]; ok {
				delete(hub.users, delUser)
				close(delUser.send)
				delUser.conn.Close()
				log.Println("Killed user", delUser.name)
				NumberOfConnections--
				// if hub has no connections, close hub
				if len(hub.users) == 0 {
					log.Println("Hub", hub.chatid, "going down")
					delete(hublist, hub.chatid)
					return
				}
			}

		case message := <-hub.broadcast:
			for user := range hub.users {
				select {
				case user.send <- message:
					log.Println("Got message in ", hub.chatid)
				default:
					close(user.send)
					delete(hub.users, user)
				}
			}

		case <-ticker.C:
			// if Chat expired, close the hub
			if !ChatExists(hub.chatid) {
				log.Println(hub.chatid, ": Closing all connections")
				for user := range hub.users {
					user.conn.Close()
					close(user.send)
					delete(hub.users, user)
					log.Println("Killed user", user.name)
				}
				delete(hublist, hub.chatid)
				return
			}
		}

	}
}
