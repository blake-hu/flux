package main

import (
	"log"
	"sync"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

func hello() {
	log.Printf("hello")
}

func CreateWebRtcConnection() (*webrtc.PeerConnection, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	se := webrtc.SettingEngine{}
	se.SetEphemeralUDPPortRange(10000, 10100)
	m := &webrtc.MediaEngine{}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		return nil, err
	}

	i := &interceptor.Registry{}

	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return nil, err
	}
	// when this is deployed it will need to be the public IP address
	se.SetNAT1To1IPs([]string{"127.0.0.1"}, webrtc.ICECandidateTypeHost)
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithSettingEngine(se), webrtc.WithInterceptorRegistry(i))
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		return nil, err
	}

	return peerConnection, nil
}

type Session struct {
	ID             string
	PeerConnection *webrtc.PeerConnection
}

type SessionManager struct {
	sessions map[string]*Session
	lock     sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (manager *SessionManager) CreateSession(id string, conn *webrtc.PeerConnection) *Session {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	session := &Session{
		ID:             id,
		PeerConnection: conn,
	}
	manager.sessions[id] = session
	return session
}

func (manager *SessionManager) GetSession(id string) (*Session, bool) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	session, exists := manager.sessions[id]
	return session, exists
}

func (manager *SessionManager) DeleteSession(id string) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	if session, exists := manager.sessions[id]; exists {
		session.PeerConnection.Close()
		delete(manager.sessions, id)
	}
}
