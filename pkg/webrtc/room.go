package webrtc

import (
	"sync"

	w "video-chat/pkg/webrtc"
	"github.com/gofiber/websocket/v2"
)

func RoomConn(c *websocket.Conn, p *w.Peers){
	var config webrtc.Configuration

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Print(err)
		return
	}

	newPeer := PeerConnectionState{
		PeerConnection: peerConnection,
		WebSocket: &ThreadSafeWrite{},
		Conn: c,
		Mutex: sync.Mutex(),
	}
}