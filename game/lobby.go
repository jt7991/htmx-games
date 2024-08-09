package game

import (
	"log"
	"madlibs-htmx/database"
	"madlibs-htmx/utils"
	"madlibs-htmx/views"
	"slices"

	"github.com/a-h/templ"
)

func GetLobbyData(roomId string) (*utils.GameData, error) {
	db, err := database.Get()
	if err != nil {
		log.Println("Error getting database: ", err)
		return nil, err
	}

	rows, err := db.Query(`WITH game_data AS (
    SELECT
        g.id AS game_id,
        r.host_id AS host_id,
        r.id AS room_id
    FROM
        rooms r
    INNER JOIN games g ON
        r.in_progress_game = g.id
    WHERE
        r.id = $1
),
user_join AS (
    SELECT
    *,
        CASE
            WHEN game_data.host_id = users.id THEN 1
            ELSE 0
        END AS is_host
    FROM
        users
    LEFT OUTER JOIN game_data ON game_data.room_id = users.room_id
    WHERE game_data.room_id = $1
)
SELECT
    id as user_id,
    username,
    is_host,
    CASE WHEN is_ready = 1 THEN 1 ELSE 0 END as is_ready
FROM
    user_join LEFT OUTER JOIN lobbies on user_join.game_id = lobbies.game_id and user_join.id = lobbies.user_id;`, roomId)

	defer rows.Close()

	var users []utils.LobbyUserData
	for rows.Next() {
		var id, username string
		var isHost, isReady int
		err = rows.Scan(&id, &username, &isHost, &isReady)
		if err != nil {
			return nil, err
		}

		users = append(users, utils.LobbyUserData{
			Id:       id,
			Username: username,
			IsHost:   isHost == 1,
			IsReady:  isReady == 1,
		})

	}
	return &utils.GameData{
		State:     "lobby",
		LobbyData: users,
		RoomId:    roomId,
	}, nil

}

func GetLobbyComponent(userId string, gameData *utils.GameData, fullPage bool) (templ.Component, error) {
	userIndex := slices.IndexFunc(gameData.LobbyData, func(u utils.LobbyUserData) bool {
		return u.Id == userId
	})

	currentUser := gameData.LobbyData[userIndex]

	return views.Lobby(gameData.LobbyData, fullPage, currentUser), nil
}

func ReadyUp(userId string, roomId string, isReady bool) error {

	gameId, err := GetInProgressGameId(roomId)

	db, err := database.Get()

	isReadyInt := 0
	if isReady {
		isReadyInt = 1
	}

	_, err = db.Exec("INSERT INTO lobbies (game_id, user_id, is_ready) VALUES ($1, $2, 1) ON CONFLICT (game_id, user_id) DO UPDATE SET is_ready = $1", gameId, userId, isReadyInt)

	if err != nil {
		log.Println("Error updating user: ", err)
		return err
	}

	return nil

}
