//hub là như một phòng chat (trung tâm chat), cho phép user trò chuyện, trao đổi data với nhau, nhằm xây dựng cơ sở hạ tầng cho real-time
package chat

//định nghĩa 1 trung tâm cho chat, có các trường mảng người dùng, luồng phát tin, luồng đăng ký, luồng hủy đăng ký
type Hub struct {
	clients map[*Client]bool //map con trỏ *client (all) giá trị bool
	broadcast chan []byte 
	//chan là một kiểu dữ liệu được sử dụng để tạo ra và quản lý các kênh (channels), một cơ chế cung cấp giao tiếp đồng bộ giữa các goroutine. Kênh cho phép các goroutine gửi và nhận dữ liệu qua nó một cách an toàn và đồng bộ.
	//nếu có hiện tượng bất đồng bộ (2 thằng chiếm chan đồng thời) thì sẽ báo lỗi
	register chan *Client //ai đó đăng ký vào phòng
	unregister chan *Client //ai đó hủy, tức là thoát khỏi phòng, hoặc nếu mở một tab mới thì nó cũng sẽ disconect cũ đi, dây là cách websocket hoạt động
}

//hàm giúp tạo phiên bản mới của hub
func NewHub() *Hub{
	return &Hub{
		broadcast: make(chan []byte), //có vẻ cái này sẽ dùng nén gói tin...
		register: make(chan *Client),
		unregister: make(chan *Client),
		clients: make(map[*Client]bool),
	}
}

//hàm run sẽ lấy hub đó và kiểm tra việc regis và unregis cà phát sóng nếu tất cả các tình huống đã được sử lý 
func (h *Hub) Run(){
	for{
		select {
		case client := <-h.register:
			h.clients[client] = true
		
			//xóa client khỏi h.clients
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
		case message := <-h.broadcast:
			for client := range h.clients{
				select{
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}