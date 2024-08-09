package utils

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func RenderComponent(c echo.Context, component templ.Component) error {
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func Redirect(c echo.Context, url string) *echo.Response {
	isHtmx := c.Request().Header.Get("HX-Request") == "true"
	if isHtmx {
		response := c.Response()
		response.Header().Set("HX-Location", url)
		return response
	} else {
		c.Redirect(302, url)
		return nil
	}
}

type CookieData struct {
	UserId string
	RoomId string
}

func GetCookieData(c echo.Context) *CookieData {
	cookieData := CookieData{}
	cookie, err := c.Cookie("user_id")

	if err != nil {
		cookieData.UserId = ""
	} else {
		cookieData.UserId = cookie.Value
	}

	cookie, err = c.Cookie("room_id")

	if err != nil {
		cookieData.RoomId = ""
	} else {
		cookieData.RoomId = cookie.Value
	}

	return &cookieData
}

func ClearRoomCookies(c echo.Context) {
	c.SetCookie(&http.Cookie{Name: "room_id", Value: "", HttpOnly: true, SameSite: http.SameSiteStrictMode, Path: "/"})
	c.SetCookie(&http.Cookie{Name: "user_id", Value: "", HttpOnly: true, SameSite: http.SameSiteStrictMode, Path: "/"})
}
