package ws

import (
	"log"
	"madlibs-htmx/game"
	"madlibs-htmx/utils"

	"github.com/a-h/templ"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/context"
)

var upgrader = websocket.Upgrader{}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	roomId string
	userId string
	send   chan *Message
}

func (c *Client) handleMessage(message *Message) {
	switch message.Action {
	case UpdateGameState:
		log.Println("USER ID", c.userId)
		userData, err := game.GetUserData(c.userId)
		if err != nil {
			log.Println("Unable to retrieve user data: ", err)
			return
		}
		message.GameData.CurrentUser = userData
		component, err := game.GetGameStateComponent(message.GameData, false)
		if err != nil {
			log.Println("Error getting game state component: ", err)
			return
		}

		htmlString, err := templ.ToGoHTML(context.Background(), component)
		if err != nil {
			log.Println("Error converting to html string: ", err)
			return
		}

		err = c.conn.WriteMessage(websocket.TextMessage, []byte(htmlString))
		if err != nil {
			log.Println("Error writing message: ", err)
		}
		return
	default:
		log.Println("Unknown action", message.Action)
		return
	}

}

func (c *Client) Write() {
	for {
		select {
		case message := <-c.send:
			log.Println("Processing message", message)
			c.handleMessage(message)
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, ctx echo.Context) {
	conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)

	if err != nil {
		log.Println(err)
		return
	}

	cookieData := utils.GetCookieData(ctx)

	if cookieData.RoomId == "" || cookieData.UserId == "" {
		log.Println("Missing room or user id")
		utils.Redirect(ctx, "/")
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan *Message), roomId: cookieData.RoomId, userId: cookieData.UserId}
	client.hub.Register <- client

	go client.Write()
}
