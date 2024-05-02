package chat

type Hub struct {
	//định nghĩa 1 trung tâm cho chat
	clients map[*Client]bool //map con trỏ *client (all) giá trị bool
	broadcast chan []byte 
	//chan là một kiểu dữ liệu được sử dụng để tạo ra và quản lý các kênh (channels), một cơ chế cung cấp giao tiếp đồng bộ giữa các goroutine. Kênh cho phép các goroutine gửi và nhận dữ liệu qua nó một cách an toàn và đồng bộ.
	//nếu có hiện tượng bất đồng bộ (2 thằng chiếm chan đồng thời) thì sẽ báo lỗi
	register chan *Client //ai đó đăng ký vào phòng
	unregister chan *Client //ai đó hủy, tức là thoát khỏi phòng, hoặc nếu mở một tab mới thì nó cũng sẽ disconect cũ đi, dây là cách websocket hoạt động
}