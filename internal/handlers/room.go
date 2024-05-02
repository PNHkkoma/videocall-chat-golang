package handlers

import (
	"fmt"
	"os"
	"time"

	w "video-chat/pkg/webrtc"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	guuid "github.com/google/uuid"
)


func RoomCreate(c *fiber.Ctx) error {
	return c.Redirect(fmt.Sprintf("/room/%s", guuid.New().String()))
}

func Room(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	if uuid == ""{
		c.Status(400)
		return nil;
	}
	uuid,suuid,_ := createOrGetRoom(uuid)
}

func RoomWebsocket(c *websocket.Conn){
	uuid := c.Params("uuid")
	if uuid == ""{
		return
	}
	_,_, room := createOrGetRoom(uuid)
	w.RoomConn(c, room.Peers)
}
//chức năng tạo (nếu chưa có) hoặc lấy phòng
func createOrGetRoom (uuid string)(String, string, *w.Room){
	//Room have type is webrtc
	
}

func RoomViewerWebsocket(c *websocket.Conn){

}

func RoomViewerConn(c *websocket.Conn, p *w.Peers){

}

//xác định thông báo websoc trông ntn để all mess di chuyển giữa websoc
type websocketMesage struct{
	Event string 'json:"event"',
	Data string 'json:"data"'
}