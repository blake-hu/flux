package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/pgvector/pgvector-go"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/pion/webrtc/v4"

	"github.com/gorilla/websocket"
)

var schema = `
CREATE TABLE IF NOT EXISTS person (
	first_name text,
	last_name text,
	email text,
	embedding vector(3)
);`

type Person struct {
	FirstName string          `db:"first_name"`
	LastName  string          `db:"last_name"`
	Email     string          `db:"email"`
	Embedding pgvector.Vector `db:"embedding"`
}

type Command string

const (
	CommandPing         Command = "ping"
	CommandPong         Command = "pong"
	CommandIceOffer     Command = "IceOffer"
	CommandIceAnswer    Command = "IceAnswer"
	CommandIceCandidate Command = "IceCandidate"
)

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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const db_user = "user"
const db_password = "password"
const db_name = "mydb"
const port = "8080"

type App struct {
	db             *sqlx.DB
	sessionManager *SessionManager
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s", err.Error())
	}
	defer db.Close()

	app := &App{db: db, sessionManager: NewSessionManager()}
	db.MustExec("CREATE EXTENSION IF NOT EXISTS vector;")
	db.MustExec(schema)

	server := http.Server{
		Addr: ":" + port,
	}
	http.HandleFunc("/", app.indexHandler)
	http.HandleFunc("/ws", app.wsHandler)

	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Failed to start server")
		os.Exit(1)
	}
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

func (app *App) handleIceOffer(payload IceOfferPayload, peerConnection *webrtc.PeerConnection, conn *websocket.Conn) error {
	desc := webrtc.SessionDescription{Type: webrtc.NewSDPType(payload.Type), SDP: payload.Sdp}
	if err := peerConnection.SetRemoteDescription(desc); err != nil {
		return err
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}

	if err := peerConnection.SetLocalDescription(answer); err != nil {
		return err
	}

	answerMsg, err := MarshalWsMessage(WsMessage{
		Command: CommandIceAnswer,
		Payload: IceAnswerPayload{
			Sdp:  peerConnection.LocalDescription().SDP,
			Type: peerConnection.LocalDescription().Type.String(),
		},
	})
	if err != nil {
		return err
	}
	conn.WriteMessage(websocket.TextMessage, []byte(answerMsg))
	return nil

}

func (app *App) wsHandler(w http.ResponseWriter, r *http.Request) {
	ws_conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection WS: %s", err.Error())
		return
	}
	defer ws_conn.Close()

	sessionId, err := gonanoid.New()
	if err != nil {
		log.Printf("Failed to create id: %s", err.Error())
	}

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	se := webrtc.SettingEngine{}
	se.SetEphemeralUDPPortRange(10000, 10100)
	// when this is deployed it will need to be the public IP address
	se.SetNAT1To1IPs([]string{"127.0.0.1"}, webrtc.ICECandidateTypeHost)
	api := webrtc.NewAPI(webrtc.WithSettingEngine(se))
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		log.Printf("Failed to create peerConnection: %s", err.Error())
		return
	}
	defer peerConnection.Close()

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			msg, err := MarshalWsMessage(WsMessage{
				Command: CommandIceCandidate,
				Payload: IceCandidatePayload{
					Candidate:     candidate.ToJSON().Candidate,
					SdpMid:        *candidate.ToJSON().SDPMid,
					SdpMLineIndex: *candidate.ToJSON().SDPMLineIndex,
				},
			})
			if err != nil {
				log.Printf("Failed to marhsal ICE candidate message: %s", err.Error())
				return
			}
			ws_conn.WriteMessage(websocket.TextMessage, []byte(msg))

		}
	})

	// need to change this to a media channel probably for video
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		log.Printf("New Data Channel %s %d", d.Label(), d.ID())
		d.OnOpen(func() {
			log.Printf("channel opened")
		})
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf("Received Message from Data Channel '%s': %s", d.Label(), string(msg.Data))
		})

		d.OnClose(func() {
			log.Printf("Data channel '%s'-'%d' closed.", d.Label(), d.ID())
		})

		d.OnError(func(err error) {
			log.Printf("Error on data channel '%s': %s", d.Label(), err.Error())
		})
	})

	for {
		_, message, err := ws_conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("Error while reading: %s", err.Error())
			}
			return
		}
		parsed_message, err := UnmarshalWsMessage(string(message))
		if err != nil {
			log.Printf("Error unmarshalling message: %s, got error: %s", message, err.Error())
		}

		switch parsed_message.Command {
		case CommandPing:
			payload, ok := parsed_message.Payload.(PingPayload)
			if !ok {
				log.Println("Invalid payload for Ping")
				return
			}
			log.Printf("Ping received: %s", payload.Data)

			response, err := MarshalWsMessage(WsMessage{
				Command: CommandPong,
				Payload: PongPayload{
					Data: payload.Data,
				},
			},
			)
			if err != nil {
				log.Printf("Error marshalling response: %s", err.Error())
				continue
			}
			if err := ws_conn.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
				log.Printf("Error sending response: %s", err.Error())
			}
		case CommandIceCandidate:
			payload, ok := parsed_message.Payload.(IceCandidatePayload)
			if !ok {
				log.Println("Invalid payload for ICE candidate")
				return
			}

			candidate := webrtc.ICECandidateInit{
				Candidate:     payload.Candidate,
				SDPMid:        &payload.SdpMid,
				SDPMLineIndex: &payload.SdpMLineIndex,
			}

			err := peerConnection.AddICECandidate(candidate)
			if err != nil {
				log.Printf("Error adding ICE candidate: %s", err.Error())
				return
			}
		case CommandIceOffer:
			payload, ok := parsed_message.Payload.(IceOfferPayload)
			if !ok {
				log.Println("Invalid payload for ICE offer")
				return
			}
			err := app.handleIceOffer(payload, peerConnection, ws_conn)
			if err != nil {
				log.Printf("Failed to handle ICE offer: %s", err.Error())
				return
			}
			app.sessionManager.CreateSession(sessionId, peerConnection)
			defer app.sessionManager.DeleteSession(sessionId)
			log.Printf("successfully created webrtc session: %s", sessionId)
		}
	}
}

func (app *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("hit index handler\n")
	people := []Person{
		{FirstName: "Max", LastName: "Glass", Email: "glass@u.northwestern.edu", Embedding: pgvector.NewVector([]float32{1, 1, 1})},
		{FirstName: "Blake", LastName: "Hu", Email: "email", Embedding: pgvector.NewVector([]float32{2, 2, 2})},
	}
	_, err := app.db.NamedExec("INSERT INTO person (first_name, last_name, email, embedding) VALUES (:first_name, :last_name, :email, :embedding)", people)
	if err != nil {
		log.Fatalf("failed to insert %s\n", err.Error())
	}

	var selected_people []Person
	app.db.Select(&selected_people, "SELECT * FROM person ORDER BY embedding <-> $1 limit 4", pgvector.NewVector([]float32{1, 1, 1}))
	fmt.Printf("people: %+v\n", selected_people)
}

// Endpoints:
// enroll new user
// - email, video
// authenticate user
// - email, video -> JWT(claim: email)
//
//			     --- python
// go server <-> |
//		         --- python

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
