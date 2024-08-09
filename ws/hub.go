package ws

import "log"

type Hub struct {
	Broadcast          chan *Message
	Register           chan *Client
	Unregister         chan *Client
	Clients            map[*Client]bool
	userIdToClientMap  map[string]*Client
	roomIdToClientsMap map[string][]*Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:          make(chan *Message),
		Register:           make(chan *Client),
		Unregister:         make(chan *Client),
		Clients:            make(map[*Client]bool),
		userIdToClientMap:  make(map[string]*Client),
		roomIdToClientsMap: make(map[string][]*Client),
	}
}

func (h *Hub) dispatchMessage(message *Message) {
	if message.ToUserId == "" {
		for _, client := range h.roomIdToClientsMap[message.ToRoomId] {
			client.send <- message
		}
	} else {
		client, ok := h.userIdToClientMap[message.ToUserId]
		if ok {
			client.send <- message
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			h.roomIdToClientsMap[client.roomId] = append(h.roomIdToClientsMap[client.roomId], client)
			log.Println("Game Id to client length", h.roomIdToClientsMap[client.roomId][0].userId)
			h.userIdToClientMap[client.userId] = client
		case client := <-h.Unregister:
			delete(h.Clients, client)
			delete(h.userIdToClientMap, client.userId)
			for i, cli := range h.roomIdToClientsMap[client.roomId] {
				if cli == client {
					h.roomIdToClientsMap[client.roomId] = append(h.roomIdToClientsMap[client.roomId][:i], h.roomIdToClientsMap[client.roomId][i+1:]...)
					break
				}
			}
		case message := <-h.Broadcast:
			h.dispatchMessage(message)
		}
	}
}
