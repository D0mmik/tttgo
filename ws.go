package main

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Name       string
	Conn       websocket.Conn
	JoinedRoom *Room
	IsPlayer   bool
}

type Room struct {
	ID         string
	Name       string
	Players    map[*Client]bool
	Spectators map[*Client]bool
}

type BroadcastRoomData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	UserCount int    `json:"userCount"`
	IsPlayer  bool   `json:"isPlayer"`
}

type JSONMessage struct {
	State string
	Value string
}

var allClients = make(map[*Client]bool)
var allRooms = make(map[string]*Room)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}
	client := newClient(conn, "username"+string(rand.IntN(100)))
	defer handleClientDisconnect(client)
	log.Print("User joined" + client.Name)

	allClients[client] = true

	broadcastStart(conn)

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Println(err)
			break
		}
		log.Print(string(message))

		var jsonMessage JSONMessage

		json.Unmarshal(message, &jsonMessage)

		if jsonMessage.State == "CREATE_GAME" {
			CreateGame(client, jsonMessage, conn)
		}

		if jsonMessage.State == "JOIN_GAME" {
			JoinGame(client, jsonMessage, conn)
		}

		if jsonMessage.State == "LEAVE_GAME" {
			LeaveGame(client)
		}

		// send message to every client in room
		if client.JoinedRoom != nil {
			for c := range allRooms[client.JoinedRoom.ID].Players {
				c.Conn.WriteMessage(websocket.TextMessage, []byte("Autobus"))
			}
		}

	}

}

func handleClientDisconnect(c *Client) {
	c.Conn.Close()

	if c.JoinedRoom != nil && allRooms[c.JoinedRoom.ID] != nil {
		//delete(allRooms, c.JoinedRoom.ID)
	}
	// maybe leave game
	delete(allClients, c)
	log.Println(allRooms)
	log.Println("User disconnected")

	broadcastUserCount()
	broadcastRooms()

}

func broadcast(v interface{}) {
	for c := range allClients {
		c.Conn.WriteJSON(v)
	}
}

func broadcastUserCount() {
	broadcast(map[string]int{"userCount": len(allClients)})
}

func broadcastRooms() {
	roomDataList := make([]BroadcastRoomData, 0, len(allRooms))

	for _, room := range allRooms {
		roomDataList = append(roomDataList, BroadcastRoomData{
			ID:        room.ID,
			Name:      room.Name,
			UserCount: len(room.Players),
		})
	}

	broadcast(map[string]interface{}{"rooms": roomDataList})
}

func broadcastStart(conn *websocket.Conn) {
	broadcastUserCount()

	roomDataList := make([]BroadcastRoomData, 0, len(allRooms))

	for _, room := range allRooms {
		roomDataList = append(roomDataList, BroadcastRoomData{
			ID:        room.ID,
			Name:      room.Name,
			UserCount: len(room.Players),
		})
	}

	conn.WriteJSON(map[string]interface{}{"rooms": roomDataList})
}

func newClient(conn *websocket.Conn, name string) *Client {
	return &Client{Conn: *conn, Name: name}
}

func newRoom(id, name string) *Room {
	return &Room{
		ID:         id,
		Name:       name,
		Players:    make(map[*Client]bool),
		Spectators: make(map[*Client]bool),
	}
}
