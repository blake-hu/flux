package main

import (
	"encoding/json"
	"fmt"
)

type Command string

const (
	CommandPing         Command = "ping"
	CommandPong         Command = "pong"
	CommandIceOffer     Command = "IceOffer"
	CommandIceAnswer    Command = "IceAnswer"
	CommandIceCandidate Command = "IceCandidate"
	CommandSetBand      Command = "setBand"
)

type SetBandPayload struct {
	Color    string `json:"color"`
	Position uint16 `json:"position"`
	Duration uint16 `json:"duration"`
}

type PingPayload struct {
	Data string `json:"data"`
}

type PongPayload struct {
	Data string `json:"data"`
}

type IceOfferPayload struct {
	Sdp  string `json:"sdp"`
	Type string `json:"type"`
}

type IceAnswerPayload = IceOfferPayload

type IceCandidatePayload struct {
	Candidate        string `json:"candidate"`
	SdpMid           string `json:"sdpMid"`
	SdpMLineIndex    uint16 `json:"sdpMLineIndex"`
	UsernameFragment string `json:"usernameFragment,omitempty"`
}

type WsMessage struct {
	Command Command     `json:"command"`
	Payload interface{} `json:"payload"`
}

func MarshalWsMessage(msg WsMessage) (string, error) {
	marshaled, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(marshaled), nil
}

func UnmarshalWsMessage(data string) (WsMessage, error) {
	var msg WsMessage
	var rawPayload json.RawMessage

	tempMsg := struct {
		Command Command          `json:"command"`
		Payload *json.RawMessage `json:"payload"`
	}{Payload: &rawPayload}

	err := json.Unmarshal([]byte(data), &tempMsg)
	if err != nil {
		return WsMessage{}, err
	}
	msg.Command = tempMsg.Command

	switch msg.Command {
	case CommandPing:
		var payload PingPayload
		if err := json.Unmarshal(*tempMsg.Payload, &payload); err != nil {
			return msg, err
		}
		msg.Payload = payload
	case CommandIceOffer:
		var payload IceOfferPayload
		if err := json.Unmarshal(*tempMsg.Payload, &payload); err != nil {
			return msg, err
		}
		msg.Payload = payload
	case CommandIceCandidate:
		var payload IceCandidatePayload
		if err := json.Unmarshal(*tempMsg.Payload, &payload); err != nil {
			return msg, err
		}
		msg.Payload = payload
	default:
		return WsMessage{}, fmt.Errorf("Unsupported command type: %s", msg.Command)
	}
	return msg, nil
}
