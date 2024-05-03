package handlers

import (
	"crypto/sha256"
	"fmt"
	"time"

	"video-chat/pkg/chat"
	w "video-chat/pkg/webrtc"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	guuid "github.com/google/uuid"
)

//return 1 room have id is google uuid
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
	return c.Render("peer", fiber.Map{
		//sprintf giúp cấu trúc lại truỗi
		"RoomWebSocketAddr": fmt.Sprintf("%s://%s/room%s/websocket", ws, c.Hostname(), uuid),
		"RoomLink":fmt.Sprintf("%s://%s/room/%s", c.Protocol(), c.Hostname(), uuid),
		"ChatWebSocketAddr":fmt.Sprintf("%s://%s/room/%s/chat/websocket", ws, c.Hostname(), uuid),
		"ViewerWebSocketAddr":fmt.Sprintf("%s://%s/room/%s/viewer/websocket", ws, c.Hostname(), uuid),
		"StreamLink":fmt.Sprintf("%s://%s/stream/%s", c.Protocol(), c.Hostname(), suuid),
		"Type":"",
	}, "layouts/main")
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
func createOrGetRoom (uuid string)(string, string, *w.Room){
	//Room have type is webrtc
	w.RoomsLock.Lock()
	defer w.RoomsLock.UnLock()
	h := sha256.New()
	h.Write([]byte(uuid)) //băm uuid
	suuid := fmt.Sprintf("%x", h.Sum(nil))

	if room := w.Rooms[uuid]; room != nil {
		if _, ok := w.Streams[suuid]; !ok{
			w.Stream[suuid] = room
		}
		return uuid, suuid, room
	}
	hub := chat.NewHub()
	p := &w.Peers{}
	p.TrackLocals = make(map[string]*webrtc.TrackLocalStaticRTP)
	room := &w.Room{
		Peers: p,
		Hub: hub,
	}
	w.Rooms[uuid] = room
	w.Stream[suuid] = room
	go hub.Run()
	return uuid, suuid, room

}

func RoomViewerWebsocket(c *websocket.Conn){
	uuid := c.Params("uuid")
	if uuid == ""{
		return
	}

	w.RoomsLock.Lock()
	if peer, ok := w.Rooms[uuid]; ok {
		w.RoomsLock.Unlock()
		RoomViewerConn(c, peer.Peers)
		return
	}
	w.RoomsLock.UnLock()
}

func RoomViewerConn(c *websocket.Conn, p *w.Peers){
	ticker := time.NewTicker(1* time.Second)
	defer ticker.Stop()
	defer c.Close()
	for {
		select{
		case <-ticker.C:
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write([]byte(fmt.Sprintf("%d",len(p.Connections))))
		}
	}
}

//xác định thông báo websoc trông ntn để all mess di chuyển giữa websoc
type websocketMesage struct{
	Event string `json:"event"`
	Data string `json:"data"`
}