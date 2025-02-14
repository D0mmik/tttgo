package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
)

func CreateGame(client *Client, jsonMessage JSONMessage, conn *websocket.Conn) {
	room := newRoom(uuid.NewString(), jsonMessage.Value)

	room.Players[client] = true
	client.JoinedRoom = room

	allRooms[room.ID] = room

	conn.WriteJSON(map[string]interface{}{"joinedRoom": BroadcastRoomData{
		ID:        room.ID,
		Name:      room.Name,
		UserCount: len(room.Players),
		IsPlayer:  true,
	}})

	broadcastRooms()
}

func JoinGame(client *Client, jsonMessage JSONMessage, conn *websocket.Conn) {
	room := allRooms[jsonMessage.Value]
	log.Println(len(room.Players))

	if len(room.Players) < 2 {
		room.Players[client] = true
		client.IsPlayer = true
	} else {
		room.Spectators[client] = true
		client.IsPlayer = false
	}

	client.JoinedRoom = room

	log.Printf("joined room")

	conn.WriteJSON(map[string]interface{}{"joinedRoom": BroadcastRoomData{
		ID:        room.ID,
		Name:      room.Name,
		UserCount: len(room.Players) + len(room.Spectators),
		IsPlayer:  client.IsPlayer,
	}})

	broadcastRooms()
}

func LeaveGame(client *Client) {

	if client.JoinedRoom == nil {
		return
	}
	room := allRooms[client.JoinedRoom.ID]

	_, playerExists := room.Players[client]

	if playerExists {
		delete(room.Players, client)
	} else {
		delete(room.Spectators, client)
	}
	broadcastRooms()
}
