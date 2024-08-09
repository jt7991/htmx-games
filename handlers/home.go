package handlers

import (
	"fmt"
	"log"
	"madlibs-htmx/common"
	"madlibs-htmx/game"
	"madlibs-htmx/utils"
	"madlibs-htmx/views"
	"madlibs-htmx/ws"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func getHome(c echo.Context) error {

	cookieData := utils.GetCookieData(c)

	if cookieData.UserId != "" && cookieData.RoomId != "" {
		log.Println("Redirecting to /game")
		utils.Redirect(c, "/game")
		return nil
	}

	component := views.Home(views.HomeViewParams{})
	return utils.RenderComponent(c, component)
}

func makeValidationErrorReadable(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required."
	case "min":
		return fmt.Sprintf("This field must be at least %s characters long.", err.Param())
	case "max":
		return fmt.Sprintf("This field must be at most %s characters long.", err.Param())
	}

	return "Invalid input"
}

type JoinGameRequest struct {
	RoomCode string `form:"room_code" validate:"required,min=3,max=10" json:"room_code"`
	Username string `form:"username" validate:"required,min=3,max=10" json:"username"`
}

func handleFormErrors(c echo.Context, err error, req JoinGameRequest) error {
	log.Printf("Validation error: %v", err)
	validationErrors := err.(validator.ValidationErrors)

	var formFields = []common.FormField{
		{FieldName: "room_code", ErrorMessage: "", Value: req.RoomCode},
		{FieldName: "username", ErrorMessage: "", Value: req.Username},
	}

	for _, e := range validationErrors {
		errorMsg := makeValidationErrorReadable(e)
		value := e.Value()
		for i, f := range formFields {
			if f.FieldName == e.Field() {
				formFields[i].ErrorMessage = errorMsg
				formFields[i].Value = value.(string)
			}
		}
	}

	component := views.HomeForm(views.HomeViewParams{ValidationResult: formFields})
	return utils.RenderComponent(c, component)
}

func joinGame(c echo.Context, hub *ws.Hub) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("form")
	})

	var req JoinGameRequest
	if err := c.Bind(&req); err != nil {
		log.Println(err.Error())
		component := views.HomeForm(views.HomeViewParams{
			MiscErrorMessage: "An unexpected error occurred. Please try again.",
		})
		return utils.RenderComponent(c, component)
	}

	if err := validate.Struct(req); err != nil {
		log.Printf("Validation error: %v", err)
		return handleFormErrors(c, err, req)
	}

	gameInfo, err := game.JoinOrCreateGame(req.RoomCode, req.Username, "madlibs")

	log.Println("gameInfo: ", gameInfo)
	if err != nil {
		log.Println(err.Error())
		component := views.HomeForm(views.HomeViewParams{
			MiscErrorMessage: err.Error(),
			ValidationResult: []common.FormField{
				{FieldName: "room_code", ErrorMessage: "", Value: req.RoomCode},
				{FieldName: "username", ErrorMessage: "", Value: req.Username},
			},
		})
		return utils.RenderComponent(c, component)
	}

	c.SetCookie(&http.Cookie{Name: "room_id", Value: gameInfo.RoomId, HttpOnly: true, SameSite: http.SameSiteStrictMode, Path: "/"})
	c.SetCookie(&http.Cookie{Name: "user_id", Value: gameInfo.UserId, HttpOnly: true, SameSite: http.SameSiteStrictMode, Path: "/"})

	gameData, err := game.GetGameData(gameInfo.RoomId, "")

	hub.Broadcast <- &ws.Message{Action: ws.UpdateGameState, GameData: gameData, ToRoomId: gameInfo.RoomId}

	utils.Redirect(c, "/game")
	return nil
}

func SetupHomeRoutes(group *echo.Group, hub *ws.Hub) {
	group.GET("", getHome)
	group.POST("/join-room", func(c echo.Context) error {
		joinGame(c, hub)
		return nil
	})
}
