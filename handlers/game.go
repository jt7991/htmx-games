package handlers

import (
	"log"
	"madlibs-htmx/game"
	"madlibs-htmx/utils"
	"madlibs-htmx/views"
	"madlibs-htmx/ws"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func getGame(c echo.Context) error {
	cookieData := utils.GetCookieData(c)

	if cookieData.RoomId == "" || cookieData.UserId == "" {
		utils.Redirect(c, "/")
		return nil
	}

	data, err := game.GetGameData(cookieData.RoomId, cookieData.UserId)

	if err != nil {
		log.Println("Error getting game data: ", err)
		return err
	}

	component, err := game.GetGameStateComponent(data, true)
	if err != nil {
		log.Println("Error getting game state component: ", err)
		return err
	}
	return utils.RenderComponent(c, component)
}

func leaveRoom(c echo.Context, hub *ws.Hub) error {
	cookieData := utils.GetCookieData(c)

	if cookieData.RoomId == "" || cookieData.UserId == "" {
		utils.Redirect(c, "/")
		return nil
	}

	err := game.LeaveRoom(cookieData.UserId)

	if err != nil {
		log.Println(err)
		return nil
	}

	utils.ClearRoomCookies(c)
	gameData, err := game.GetGameData(cookieData.RoomId, "")

	if err != nil {
		log.Println("Error getting game data: ", err)
		return err
	}

	hub.Broadcast <- &ws.Message{Action: ws.UpdateGameState, ToRoomId: cookieData.RoomId, GameData: gameData}

	utils.Redirect(c, "/")
	return nil
}

type ReadyUpRequest struct {
	Ready bool `form:"ready" validate:"required"`
}

func readyUp(c echo.Context, hub *ws.Hub) error {

	cookieData := utils.GetCookieData(c)

	if cookieData.RoomId == "" || cookieData.UserId == "" {
		utils.Redirect(c, "/")
		return nil
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("form")
	})

	var req ReadyUpRequest
	if err := c.Bind(&req); err != nil {
		log.Println("Error binding request: ", err)
		component := views.Toast(err.Error(), views.ToastErrorVariant)
		return utils.RenderComponent(c, component)
	}

	if err := validate.Struct(req); err != nil {
		log.Printf("Validation error: %v", err)
		component := views.Toast(err.Error(), views.ToastErrorVariant)
		return utils.RenderComponent(c, component)
	}
	err := game.ReadyUp(cookieData.UserId, cookieData.RoomId, req.Ready)

	gameData, err := game.GetGameData(cookieData.RoomId, "")

	if err != nil {
		log.Println("Error getting game data: ", err)
		return err
	}

	hub.Broadcast <- &ws.Message{Action: ws.UpdateGameState, ToRoomId: cookieData.RoomId, GameData: gameData}
	return nil
}

func wsConnect(c echo.Context, hub *ws.Hub) {
	ws.ServeWs(hub, c)
}

func SetupGameRoutes(group *echo.Group, hub *ws.Hub) {
	group.GET("", getGame)
	group.GET("/ws", func(c echo.Context) error {
		wsConnect(c, hub)
		return nil
	})

	group.POST("/leave-room", func(c echo.Context) error {
		leaveRoom(c, hub)
		return nil
	})

	group.POST("/ready-up", func(c echo.Context) error {
		readyUp(c, hub)
		return nil
	})
}
