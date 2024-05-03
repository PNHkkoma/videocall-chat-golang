package handlers

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)



func Stream( c *fiber.Ctx) error{
	suuid := c.Params("suuid")
	if suuid == ""{
		c.Status(400)
		return nil
	}

	ws := "ws"
	if os.Getenv("ENVIROMENT") == "PRODUCTION" {
		ws = "wsss"
	}

	w.RoomsLock.Lock()
	if _,ok := w.Stream[suuid]; ok {
		w.RoomsLock.Unlock()
		return c.Render("stream", fiber.Map{
			"StreamWebsocketAddr": fmt.Sprintf("%s://%s/stream/%s/websocket", ws, c.Hostname(), suuid),
			"ChatWebsocketAddr": fmt.Sprintf("%s://%s/stream/%s/chat/webdocket", ws, c.Hostname(), suuid),
			"ViewerWebsocketAddr":fmt.Sprintf("%s://%s/stream/%s/viewer/webdocket", ws, c.Hostname(), suuid),
			"Type":"stream",
		}, "layouts/main")
	}
	w.RoomsLock.Unlock()
	return c.Render("stream", fiber.Map{
		"NoStream":"true",
		"Leave":"true",
	},"layouts/main")
}

func StreamWebsocket(c * fiber.Ctx) error {

}

func StreamViewerWebsocket()