package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
)

func CreateGame(client *Client, jsonMessage JSONMessage, conn *websocket.Conn) {
	room := newRoom(uuid.NewString(), jsonMessage.Value)
	client.Ready = false

	room.Players[client.ID] = client
	client.JoinedRoom = room

	allRooms[room.ID] = room
	client.IsPlayer = true
	client.IsLeader = true

	conn.WriteJSON(map[string]interface{}{"joinedRoom": BroadcastRoomData{
		ID:        room.ID,
		Name:      room.Name,
		UserCount: len(room.Players),
		IsPlayer:  client.IsPlayer,
		IsLeader:  client.IsLeader,
	}})

	broadcastRooms()
}

func JoinGame(client *Client, jsonMessage JSONMessage, conn *websocket.Conn) {
	room := allRooms[jsonMessage.Value]
	client.Ready = false

	if len(room.Players) < 2 {
		room.Players[client.ID] = client
		client.IsPlayer = true
	} else {
		room.Spectators[client.ID] = client
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

	_, playerExists := room.Players[client.ID]

	if playerExists {
		delete(room.Players, client.ID)
	} else {
		delete(room.Spectators, client.ID)
	}
	broadcastRooms()
}
