package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/pgvector/pgvector-go"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264writer"
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection WS: %s", err.Error())
		return
	}
	defer ws.Close()

	sessionId, err := gonanoid.New()
	if err != nil {
		log.Printf("Failed to create id: %s", err.Error())
	}

	peerConnection, err := CreateWebRtcConnection()
	if err != nil {
		log.Printf("Failed to create peer connection: %s", err.Error())
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
			ws.WriteMessage(websocket.TextMessage, []byte(msg))

		}
	})

	h264File, err := h264writer.New("/video/output.mp4")
	if err != nil {
		panic(err)
	}

	peerConnection.OnTrack(func(track *webrtc.TrackRemote, recv *webrtc.RTPReceiver) {
		codec := track.Codec()
		log.Printf("New Track %s %s", track.Kind().String(), track.ID())
		if strings.EqualFold(codec.MimeType, webrtc.MimeTypeVP8) {
			fmt.Println("Got VP8 track, saving to disk as output.ivf")
			saveToDisk(h264File, track)
		}
	})

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("Error while reading: %s", err.Error())
			}
			return
		}
		parsed_message, err := UnmarshalWsMessage(string(message))
		if err != nil {
			log.Printf("Error unmarshalling message: %s, got error: %s", message, err.Error())
			return
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
			})

			if err != nil {
				log.Printf("Error marshalling response: %s", err.Error())
				return
			}
			if err := ws.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
				log.Printf("Error sending response: %s", err.Error())
				return
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
			err := app.handleIceOffer(payload, peerConnection, ws)
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

func saveToDisk(i media.Writer, track *webrtc.TrackRemote) {
	defer func() {
		if err := i.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := i.WriteRTP(rtpPacket); err != nil {
			fmt.Println(err)
			return
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
