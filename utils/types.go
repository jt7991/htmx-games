package utils

type LobbyUserData struct {
	Id       string
	Username string
	IsHost   bool
	IsReady  bool
}

type GameData struct {
	State       string
	RoomId      string
	GameId      string
	LobbyData   []LobbyUserData
	CurrentUser *LobbyUserData
}
