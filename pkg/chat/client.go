// hàm định nghĩa và các chức năng của client
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
	space = []byte{' '}
)

//cần 2 thứ đó là kích thước bộ đệm đột kích và
var upgrader = websocket.FastHTTPUpgrader {
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

type Client struct{
	//tạo biến client
	Hub *Hub //có con trỏ tương ứng với hub
	Conn *websocket.Conn //connect trỏ tới websocket
	Send chan []byte //channel send dạng byte? chăc đây là tin gửi đi khi đc nén
}

//Pump có tác dụng kiểu sử lý dữ liệu đến và đi, điều phối dữ liệu, bộ lọc và sử lý ngoại lệ
func (c *Client) readPump(){
	//defer kiểu chạy cuối ấy
	defer func(){
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize) //sets the maximum size in bytes for a message read from the peer
	c.Conn.SetReadDeadline(time.Now().Add(pongWait)) //sets the write deadline(giới hạn thời gian) on the underlying(c) network connection
	//SetPongHandler đặt trình xử lý cho các tin nhắn pong, ping là 1 mess từ 1 endpoint tơi endpoint khác, còn pong là tin nhắn phản hồi lại xác nhận kết  nối vẫn hoạt động và endpoint vẫn sẵng sàng để chuyền tải dữ liệu
	c.Conn.SetPongHandler(func(string) error {c.Conn.SetReadDeadline(time.Now().Add(pongWait));		return nil})
	//duyệt qua các mess
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
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.Hub.broadcast <- message
	}
}

func (c *Client) writePump(){
	//đánh dấu time gửi tin của peer hiển thị trong room chat
	ticker := time.NewTicker(pingPeriod)
	defer func(){
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select{
			//khi nhận được một số tin nhắn
		case message, ok := <-c.Send:
			//nếu nó đc gửi thì set thời gian đợi
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok{
				return
			}
			//viết theo nguyên tắc của nextWrite
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil{
				return 
			}
			w.Write(message)
			n := len(c.Send)
			//xuống dòng sau khi write xong content message
			for i:=0; i<n;i++{
				w.Write(newline)
				w.Write(<-c.Send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <- ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return 
			}
		}
	}
}

//kết nối trò truyện ngang hàng
func PeerChatConn(c *websocket.Conn, hub *Hub) {
	client := &Client{Hub: hub, Conn: c, Send: make(chan []byte,256)}
	client.Hub.register <- client
	
	go client.writePump()
	client.readPump()
}