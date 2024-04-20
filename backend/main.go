package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pgvector/pgvector-go"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	// "github.com/pion/interceptor"
	// "github.com/pion/interceptor/pkg/intervalpli"
	// "github.com/pion/webrtc/v4"

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

//	type PingMessage struct {
//		Ping string `json:"ping"`
//	}
//
//	type PongMessage struct {
//		Pong string `json:"pong"`
//	}
//
//	type WsMessage struct {
//		Type string      `json:"type"`
//		Data interface{} `json:"data"`
//
// }

type Command string

const (
	CommandPing Command = "ping"
	CommandPong Command = "pong"
)

type WsMessage struct {
	Command Command     `json:"command"`
	Payload interface{} `json:"payload"`
}
type PingPayload struct {
	Data string `json:"data"`
}
type PongPayload struct {
	Data string `json:"data"`
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
	db *sqlx.DB
}

func main() {
	db, err := sqlx.Connect("postgres", fmt.Sprintf("host=db user=%s dbname=%s password=%s sslmode=disable", db_user, db_name, db_password))
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s", err.Error())
	}
	defer db.Close()

	// m := &webrtc.MediaEngine{}
	//
	// if err := m.RegisterCodec(webrtc.RTPCodecParameters{
	// 	RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
	// 	PayloadType:        96,
	// }, webrtc.RTPCodecTypeVideo); err != nil {
	// 	log.Fatalf("Failed to register VP8 coded: %s", err.Error())
	// }
	//
	// i := &interceptor.Registry{}
	//
	// intervalPliFactory, err := intervalpli.NewReceiverInterceptor()
	// if err != nil {
	// 	log.Fatalf("Failed to create NewReceiverInterceptor: %s", err.Error())
	// }
	// i.Add(intervalPliFactory)
	//
	// if err = webrtc.RegisterDefaultInterceptors(m, i); err != nil {
	// 	log.Fatalf("Failed to register interceptors: %s", err.Error())
	// }
	//
	// api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i))
	//
	// config := webrtc.Configuration{
	// 	ICEServers: []webrtc.ICEServer{
	// 		{
	// 			URLs: []string{"stun:stun.l.google.com:19302"},
	// 		},
	// 	},
	// }
	//
	// peerConnection, err := api.NewPeerConnection(config)
	// if err != nil {
	// 	log.Fatalf("Failed to connect to peer: %s", err.Error())
	// }
	//
	// if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
	// 	log.Fatalf("Failed to add video transceiver: %s", err.Error())
	// }
	//
	// peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	// 	codec := track.Codec()
	// 	if strings.EqualFold(codec.MimeType, webrtc.MimeTypeVP8) {
	// 		log.Print("Got VP8 track")
	// 		packet, _, err := track.ReadRTP()
	// 		if err != nil {
	// 			log.Printf("failed to packet")
	// 		}
	// 		log.Printf("unmarshalled packet: %s", packet.String())
	// 	}
	// })
	//
	// peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
	// 	log.Printf("Connection state changed %s", connectionState.String())
	//
	// 	if connectionState == webrtc.ICEConnectionStateConnected {
	// 		log.Printf("Connection started")
	// 	} else if connectionState == webrtc.ICEConnectionStateFailed || connectionState == webrtc.ICEConnectionStateClosed {
	// 		log.Printf("Connectiion closed")
	//
	// 		if closeErr := peerConnection.Close(); closeErr != nil {
	// 			log.Fatalf("Failed to close peer connection: %s", closeErr.Error())
	// 		}
	// 	}
	// })

	app := &App{db: db}
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

	fmt.Printf("Sever listening on port %s\n", port)
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
	default:
		return WsMessage{}, fmt.Errorf("Unsupported command type: %s", msg.Command)
	}
	return msg, nil
}

func (app *App) wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection WS: %s", err.Error())
		return
	}

	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
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
			if err := c.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
				log.Printf("Error sending response: %s", err.Error())
			}

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
