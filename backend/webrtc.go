package main

import (
	"log"

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
