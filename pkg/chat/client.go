// cần 3 chức năng ở đây đó là
package chat

import (
	"bytes"
	"log"
	"time"

	"github.com/fasthttp/websocket"
)

const(
	//cố định các thời gian, có thể bỏ qua nhưng đó là giới hạn khi có bug
	writeWait = 10 * time.Second 
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10 
	maxMessageSize =512 //kích thước tin tối đa
)

var (
	newline = []byte{'\n'}
	space = []byte{''}
)

//cần 2 thứ đó là kích thước bộ đệm đột kích và
var upgrader = websocket.FastHTTPUpgrader (
	ReadBufferSize: 1024,
	WriteBufferSize: 1024
)

type Client struct{
	//tạo biến client
	Hub *Hub //có con trỏ tương ứng với hub
	Conn *websocket.Conn //connect trỏ tới websocket
	Send chan []byte //channel send dạng byte?
}

func (c *Client) readPump(){
	defer func(){
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {c.Conn.SetReadDeadline(time.Now().Add(pongWait));		return nil})
	d=//duyệt qua các mess
	for{
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			//trường hợp lỗi phổ biến: websoc bị đóng nên muốn sử lý đối với lỗi đóng ko mong muốn, nếu ko sẽ ko biết cái lỗi gì đang sảy ra
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure){
				log.Printf("error: %v", err)
			}
			break;
		}
		//cắt bớt khoảng trắng của mess
		message = byte.TrimSpace[bytes.Replace(message, newline, space, -1)]
		c.Hub.broadcast <- message
	}
}

func (c *Client) writePump(){

}

//kết nối trò truyện ngang hàng
func PeerChatConn(c *websocket.Conn, hub *Hub) {
	client := &Client{Hub: hub, Conn: c, Send: make(chan []byte,256)}
	client.Hub.register <- client
	
	go client.writePump()
	client.readPump()
}