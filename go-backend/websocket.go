package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Command string

const (
	CommandIceOffer          Command = "IceOffer"
	CommandIceAnswer         Command = "IceAnswer"
	CommandIceCandidate      Command = "IceCandidate"
	CommandSetBandColor      Command = "setBandColor"
	CommandReadyForBandColor Command = "readyForBandColor"
	ColorCommandAck          Command = "colorCommandAck"
)

type ReadyForBandColorPayload struct {
	Email string `json:"email"`
}

type ColorCommandAckPayload struct {
	BackgroundColor string    `json:"backgroundColor"`
	StripColor      string    `json:"stripColor"`
	StripPosition   string    `json:"stripPosition"`
	Timestamp       time.Time `json:"timestamp"`
	Index           int       `json:"index"`
}

type SetBandColorPayload struct {
	BackgroundColor string `json:"backgroundColor"`
	StripColor      string `json:"stripColor"`
	StripPosition   uint16 `json:"stripPosition"`
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
	case CommandReadyForBandColor:
		var payload ReadyForBandColorPayload
		if err := json.Unmarshal(*tempMsg.Payload, &payload); err != nil {
			return msg, err
		}
		msg.Payload = payload
	case ColorCommandAck:
		var payload ColorCommandAckPayload
		if err := json.Unmarshal(*tempMsg.Payload, &payload); err != nil {
			return msg, err
		}
		msg.Payload = payload

	default:
		return WsMessage{}, fmt.Errorf("unsupported command type: %s", msg.Command)
	}
	return msg, nil
}
