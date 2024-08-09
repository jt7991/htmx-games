package game

import (
	"database/sql"
	"fmt"
	"log"
	"madlibs-htmx/common"
	"madlibs-htmx/database"
	"madlibs-htmx/utils"

	"github.com/a-h/templ"
)

type GameName string

const (
	MADLIBS GameName = "madlibs"
)

type BaseGameInfo struct {
	RoomId string
	UserId string
}

func GetGameState(roomId string) (string, error) {
	return "lobby", nil
}

func GetGameData(roomId string, userId string) (*utils.GameData, error) {
	state, err := GetGameState(roomId)
	if err != nil {
		return nil, err
	}

	if state == "lobby" {
		data, err := GetLobbyData(roomId)
		if err != nil {
			return nil, err
		}

		if userId != "" {
			data.CurrentUser, err = GetUserData(userId)
		}

		return data, err
	}

	return nil, nil
}

func GetGameStateComponent(gameData *utils.GameData, fullPage bool) (templ.Component, error) {
	if gameData.State == "lobby" {
		return GetLobbyComponent(gameData.CurrentUser.Id, gameData, fullPage)
	}
	return nil, nil
}

func GetUserData(userId string) (*utils.LobbyUserData, error) {
	db, err := database.Get()
	if err != nil {
		return nil, err
	}

	var id, username, roomId string
	err = db.QueryRow("SELECT id, username, room_id FROM users WHERE id = ?", userId).Scan(&id, &username, &roomId)

	if err != nil {
		return nil, err
	}

	var roomHostId string

	err = db.QueryRow("SELECT host_id FROM rooms WHERE id = ?", roomId).Scan(&roomHostId)

	return &utils.LobbyUserData{
		Id:       id,
		Username: username,
		IsHost:   id == roomHostId,
	}, nil
}

func GetUsersInRoom(roomId string) ([]utils.LobbyUserData, error) {
	db, err := database.Get()
	if err != nil {
		return nil, err
	}

	var hostId string
	err = db.QueryRow("SELECT host_id FROM rooms WHERE id = $1", roomId).Scan(&hostId)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no room found with id %s", roomId)
		}
		return nil, err
	}

	var users []utils.LobbyUserData

	rows, err := db.Query("SELECT id, username FROM users WHERE room_id = $1", roomId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, username string
		err = rows.Scan(&id, &username)
		if err != nil {
			return nil, err
		}

		users = append(users, utils.LobbyUserData{
			Id:       id,
			Username: username,
			IsHost:   id == hostId,
		})
	}

	return users, nil
}

func JoinOrCreateGame(roomCode string, username string, game GameName) (*BaseGameInfo, error) {
	roomId, err := GetRoomId(roomCode, game)

	if err != nil {
		return nil, err
	}

	if roomId == "" {
		return CreateNewRoom(roomCode, username)
	}

	userInGame, err := IsUserInGame(roomCode, username, game)

	if err != nil {
		log.Println("Error checking if user is in game: ", err)
		return nil, common.UnknownError
	}

	if userInGame {
		return nil, common.UsernameTakenError
	}

	return JoinInProgressGame(roomId, username, game)

}

func doesRoomHaveHost(roomId string) (bool, error) {
	db, err := database.Get()
	if err != nil {
		return false, err
	}

	var hostId string
	err = db.QueryRow("SELECT host_id FROM rooms inner join users on rooms.host_id = users.id WHERE rooms.id = ?", roomId).Scan(&hostId)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return hostId != "", nil
}

func addHostToRoom(roomId string, userId string) error {
	db, err := database.Get()
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE rooms SET host_id = ? WHERE id = ?", userId, roomId)

	if err != nil {
		return err
	}

	log.Println("Added host to room")
	return nil
}

func LeaveRoom(userId string) error {
	log.Println("Leaving room")
	db, err := database.Get()
	if err != nil {
		log.Println("Error getting database: ", err)
		return err
	}
	_, err = db.Exec("DELETE FROM users WHERE id = ?", userId)
	if err != nil {
		log.Println("Error deleting user: ", err)
		return err
	}

	return nil
}

func CreateNewRoom(roomCode string, username string) (*BaseGameInfo, error) {
	db, err := database.Get()
	if err != nil {
		return nil, err
	}
	transaction, err := db.Begin()

	if err != nil {
		return nil, err
	}

	userId := database.GetRandomId()
	roomId := database.GetRandomId()
	gameId := database.GetRandomId()

	_, err = transaction.Exec("INSERT INTO rooms (id, room_code, in_progress_game, host_id) VALUES (?, ?, ?, ?)", roomId, roomCode, gameId, userId)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	_, err = transaction.Exec("INSERT INTO users (id, username, room_id) VALUES (?, ?, ?)", userId, username, roomId)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	_, err = transaction.Exec("INSERT INTO games (id, room_id, game_name) VALUES (?, ?, ?)", gameId, roomId, "madlibs")
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	err = transaction.Commit()

	if err != nil {
		return nil, err
	}

	return &BaseGameInfo{
		RoomId: roomId,
		UserId: userId,
	}, nil
}

func JoinInProgressGame(roomId string, username string, game GameName) (*BaseGameInfo, error) {
	db, err := database.Get()
	if err != nil {
		return nil, err
	}

	userId := database.GetRandomId()

	if err != nil {
		return nil, err
	}

	_, err = db.Exec("INSERT INTO users (id, username, room_id) VALUES (?, ?, ?)", userId, username, roomId)

	if err != nil {
		return nil, err
	}

	roomHasHost, err := doesRoomHaveHost(roomId)

	if err != nil {
		return nil, err
	}

	if !roomHasHost {
		err = addHostToRoom(roomId, userId)
	}

	if err != nil {
		return nil, err
	}

	return &BaseGameInfo{
		RoomId: roomId,
		UserId: userId,
	}, nil
}

func GetRoomId(roomCode string, game GameName) (string, error) {
	db, err := database.Get()
	if err != nil {
		return "", err
	}

	res, err := db.Query("SELECT id FROM rooms WHERE room_code = ? ", roomCode, game)

	if err != nil {
		return "", err
	}

	defer res.Close()
	row_exists := res.Next()

	if !row_exists {
		return "", nil
	}

	var id string
	err = res.Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func IsUserInGame(roomCode string, username string, game GameName) (bool, error) {
	db, err := database.Get()
	if err != nil {
		return false, err
	}

	res, err := db.Query("SELECT * FROM rooms inner join users on rooms.id = users.room_id WHERE rooms.room_code = ? AND users.username = ?", roomCode, username)

	if err != nil {
		return false, err
	}

	defer res.Close()
	return res.Next(), nil
}

func GetInProgressGameId(roomId string) (string, error) {
	db, err := database.Get()
	if err != nil {
		return "", err
	}

	var id string
	err = db.QueryRow("SELECT in_progress_game FROM rooms WHERE id = ?", roomId).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}
