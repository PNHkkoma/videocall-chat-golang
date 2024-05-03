package handlers

import (
	"video-chat/pkg/chat"
	w "video-chat/pkg/webrtc"
	"github.com/gofiber/websocket/v2"
)

//sử lý trò truyện
func RoomChat(c *firber.Ctx) error{
	return c.Render("chat", fiber.Map(), "layouts/main")
}

//kết nối phòng chat
func RoomChatWebsocket(c *websocket.Conn){
	uuid := c.Params("uuid")
	if uuid == ""{
		return
	}
	w.RoomsLock.Lock()
	room := w.Rooms[uuid]
	w.RoomsLock.Unlock()
	if room == nil {
		return
	}
	if room.Hub ==nil{
		return
	}
	chat.PeerChatConn(c.Conn, room.Hub)

}

func StreamChatWebsocket(c *websocket.Conn){
	suuid := c.Params("suuid")
	if suuid == "" {
		return
	}
	w.RoomsLock.Lock()
	if stream, ok := w.Streams[suuid]; ok{
		w.RoomsLock.Unlock()
		if stream.Hub == nil {
			hub := chat.NewHub()
			stream.Hub = hub
			go hub.Run()
		}
		chat.PeerChatConn(c.Conn, stream.Hub)
		return 
	}
	w.RoomsLock.Unlock()
}