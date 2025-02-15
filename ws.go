package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"

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
	ID         string
	Name       string
	Conn       *websocket.Conn
	JoinedRoom *Room
	IsPlayer   bool
	IsLeader   bool
	Ready      bool
	IsX        bool
}

type Room struct {
	ID         string
	Name       string
	Players    map[string]*Client
	ReadyCount int
	Spectators map[string]*Client
	Game
}

type Game struct {
	GameIsRunning bool     `json:"gameIsRunning"`
	TurnCount     uint8    `json:"turnCount"`
	Blocks        [9]uint8 `json:"blocks"`
	XPlays        bool     `json:"xPlays"`
}

type BroadcastRoomData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	UserCount int    `json:"userCount"`
	IsPlayer  bool   `json:"isPlayer"`
	IsLeader  bool   `json:"isLeader"`
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

		if jsonMessage.State == "READY" {
			if !client.IsPlayer {
				return
			}

			client.Ready = true
			client.JoinedRoom.ReadyCount += 1

			for _, c := range allRooms[client.JoinedRoom.ID].Players {
				c.Conn.WriteJSON(map[string]int{"readyCount": client.JoinedRoom.ReadyCount})
			}
		}

		if jsonMessage.State == "START_GAME" {
			client.JoinedRoom.Game.GameIsRunning = true
			client.JoinedRoom.TurnCount = 1

			r := rand.IntN(2) + 1
			i := 1

			for _, p := range client.JoinedRoom.Players {
				p.IsX = i == r
				client.JoinedRoom.XPlays = i == r
				i += 1
			}

			for _, c := range allRooms[client.JoinedRoom.ID].Players {
				c.Conn.WriteJSON(map[string]bool{"isX": c.IsX})
				c.Conn.WriteJSON(map[string]Game{"game": client.JoinedRoom.Game})
			}

			for _, c := range allRooms[client.JoinedRoom.ID].Spectators {
				c.Conn.WriteJSON(map[string]Game{"game": client.JoinedRoom.Game})
			}

		}

		if jsonMessage.State == "GAME_MOVE" {
			if !client.IsPlayer {
				return
			}

			pos, _ := strconv.Atoi(jsonMessage.Value)
			client.JoinedRoom.TurnCount += 1
			client.JoinedRoom.XPlays = !client.JoinedRoom.XPlays
			if client.IsX {
				client.JoinedRoom.Game.Blocks[pos] = 1
			} else {
				client.JoinedRoom.Game.Blocks[pos] = 2
			}

			for _, c := range allRooms[client.JoinedRoom.ID].Players {
				c.Conn.WriteJSON(map[string]Game{"game": client.JoinedRoom.Game})
			}

			for _, c := range allRooms[client.JoinedRoom.ID].Spectators {
				c.Conn.WriteJSON(map[string]Game{"game": client.JoinedRoom.Game})
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
	return &Client{ID: uuid.NewString(), Conn: conn, Name: name}
}

func newRoom(id, name string) *Room {
	return &Room{
		ID:         id,
		Name:       name,
		Players:    make(map[string]*Client),
		Spectators: make(map[string]*Client),
	}
}
