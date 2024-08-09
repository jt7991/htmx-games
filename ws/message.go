package ws

import "madlibs-htmx/utils"

type Action string

const (
	UpdateGameState Action = "updateGameState"
)

type Message struct {
	Action   Action
	GameData *utils.GameData
	UserData *utils.LobbyUserData
	ToUserId string
	ToRoomId string
}
