package webrtc

import (
	"sync"
	"video-chat/pkg/chat"
)

type Room struct{
	//tạo ra cấu trúc của 1 room gồm có
	Peers *Peers //con trỏ, tức là room sẽ gồm các peer
	Hub *chat.Hub //phòng chat (dịch là trung tâm = center) hay công cụ của socket á
}

type Peers struct{
	//cấu trúc của các peer
	ListLock sync.RWMutex //có sync là biến đồng bộ đảm bảo tính nhất quán khi truy cập và chỉnh sửa Connections và TrackLocals, sync.RWMutex là một loại mutex trong Go, cho phép đồng bộ hóa truy cập đồng thời từ nhiều goroutines nhưng vẫn đảm bảo tính nhất quán.
	//mutex là một khái niệm liên quan đến đồng bộ hóa, được sử dụng để đảm bảo chỉ một goroutine có thể truy cập vào một biến hoặc một phần của mã vào một thời điểm.
	/*
	Có hai loại mutex phổ biến trong Go:

	sync.Mutex: Đây là loại mutex tiêu chuẩn trong Go.
	 Nó cung cấp hai phương thức chính là Lock() để khóa mutex và Unlock() để mở khóa mutex.
	
	sync.RWMutex: Đây là một biến thể của Mutex 
	được sử dụng khi có nhiều goroutine muốn đọc dữ liệu một cách đồng thời
	 và chỉ một goroutine được phép thực hiện ghi. 
	 RWMutex cung cấp các phương thức RLock() để khóa cho việc đọc (read lock) 
	 và RUnlock() để mở khóa cho việc đọc. 
	 Ngoài ra, nó cũng cung cấp Lock() và Unlock() để thực hiện việc ghi (write).
	*/
	Connections []PeerConnectionState //list connect of peers
	TrackLocals map[string]*webrtc.TrackLocalStaticRTP // Được sử dụng để biểu diễn một luồng dữ liệu định tuyến thời gian thực (Real-Time Transport Protocol - RTP) cục bộ và không thay đổi (static).
}
func (p *Peers) DispatchKeyFrames(){

}