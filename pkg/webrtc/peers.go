package webrtc

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"video-chat/pkg/chat"

	"github.com/gofiber/websocket/v2"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

var (
	RoomsLock sync.RWMutex //đánh dấu Room đã bị lock ko dùng đc 2 biến dưới
	Rooms     map[string]*Room
	Streams   map[string]*Room
)

var (
	//config gì đó cần làm rõ
	turnConfig = webrtc.Configuration{
		ICETransportPolicy: webrtc.ICETransportPolicyRelay,
		ICEServers: []webrtc.ICEServer{
			{

				URLs: []string{"stun:turn.localhost:3478"},
			},
			{

				URLs: []string{"turn:turn.localhost:3478"},

				Username: "akhil",

				Credential:     "sharma",
				CredentialType: webrtc.ICECredentialTypePassword,
			},
		},
	}
)

//định nghĩa 1 Room gồm các Peers tham gia và hub chat được kết nối
type Room struct{
	Peers *Peers
	Hub *chat.Hub 
}

type Peers struct{
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

type PeerConnectionState struct {
	PeerConnection *webrtc.PeerConnection
	Websocket      *ThreadSafeWriter
}

type ThreadSafeWriter struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
}

func (t *ThreadSafeWriter) WriteJSON(v interface{}) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	return t.Conn.WriteJSON(v)
}

func (p *Peers) AddTrack (t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP{
	p.ListLock.Lock()
	defer func ()  {
		p.ListLock.Unlock()
		p.SignalPeerConnections()
	}()

	trackLocal, err :=webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil{
		log.Printf(err.Error())
		return nil
	}
	p.TrackLocals[t.ID()] = trackLocal
	return trackLocal
}

func (p *Peers) RemoveTrack (t *webrtc.TrackLocalStaticRTP){
	p.ListLock.Lock()
	defer func(){
		p.ListLock.Unlock()
		p.SignalPeerConnections()
	}()
	delete(p.TrackLocals, t.ID())
}

func (p *Peers) SignalPeerConnections(){
	p.ListLock.Lock()
	defer func(){
		p.ListLock.Unlock()
		p.DispatchKeyFrame()
	}()

	attemptSync := func() (tryAgain bool){
		for i := range p.Connections {
			if p.Connections[i].PeerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed{
				p.Connections = append(p.Connections[:i], p.Connections[i+1:]...)
				log.Printf("a", p.Connections)
				return true
			}

			existingSenders := map[string]bool{}
			for _, sender := range p.Connections[i].PeerConnection.GetSenders(){
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				if _, ok := p.TrackLocals[sender.Track().ID()]; !ok {
					if err := p.Connections[i].PeerConnection.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}
			for _, receiver := range p.Connections[i].PeerConnection.GetReceivers(){
				if receiver.Track() == nil {
					continue
				}
				existingSenders[receiver.Track().ID()] = true
			}

			for trackID := range p.TrackLocals{
				if _, ok := existingSenders[trackID]; !ok{
					if _, err := p.Connections[i].PeerConnection.AddTrack(p.TrackLocals[trackID]); err!=nil {
						return true
					}
				}
			}

			offer, err := p.Connections[i].PeerConnection.CreateOffer(nil)
			if err != nil {
				return true
			}

			if err = p.Connections[i].PeerConnection.SetLocalDescription(offer); err != nil {
				return true
			}

			offerString, err := json.Marshal(offer)
			if err != nil {
				return true
			}

			if err = p.Connections[i].Websocket.WriteJSON(&websocketMessage{
				Event: "offer",
				Data:  string(offerString),
			}); err != nil {
				return true
			}
		}
		return
	}
	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			go func() {
				time.Sleep(time.Second * 3)
				p.SignalPeerConnections()
			}()
			return
		}

		if !attemptSync() {
			break
		}
	}
}

func (p *Peers) DispatchKeyFrame(){
	p.ListLock.Lock()
	defer p.ListLock.Unlock()

	for i := range p.Connections {
		for _, receiver := range p.Connections[i].PeerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = p.Connections[i].PeerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}

type wwebsockerMessage struct{
	Event string `json:"event"`
	Data string `json:"data"`
}