package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

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
CREATE TABLE IF NOT EXISTS registered_user (
	email TEXT PRIMARY KEY,
	embedding vector(3) NOT NULL
);`

type User struct {
	Email     string          `db:"email"`
	Embedding pgvector.Vector `db:"embedding"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const PORT = "8080"
const COLOR_CHALLENGES = 10
const EMBEDDING_THRESHOLD = 0.5
const INFERENCE_BACKEND_URL = "http://inference-backend:5000"

type App struct {
	db             *sqlx.DB
	sessionManager *SessionManager
}

func main() {
	fmt.Print("Starting go backend version 1.2\n")
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
		Addr: ":" + PORT,
	}
	http.HandleFunc("/enroll", app.enrollHandler)
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

	folderPath := "./files/video"
	os.MkdirAll(folderPath, os.ModePerm)
	h264File, err := h264writer.New(fmt.Sprintf("%s/%s.mp4", folderPath, sessionId))
	if err != nil {
		panic(err)
	}

	peerConnection.OnTrack(func(track *webrtc.TrackRemote, recv *webrtc.RTPReceiver) {
		codec := track.Codec()
		log.Printf("New Track %s %s", track.Kind().String(), track.ID())
		if strings.EqualFold(codec.MimeType, webrtc.MimeTypeVP8) {
			fmt.Printf("Got VP8 track, saving to disk as %s.mp4\n", sessionId)
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

		var email string

		switch parsed_message.Command {
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
			app.sessionManager.CreateSession(sessionId, peerConnection, ws)
			defer app.sessionManager.DeleteSession(sessionId)
			log.Printf("successfully created webrtc session: %s", sessionId)
		case CommandReadyForBandColor:
			payload, ok := parsed_message.Payload.(ReadyForBandColorPayload)
			if !ok {
				log.Println("Invalid payload for ReadyForBandColor")
				return
			}
			session, exists := app.sessionManager.GetSession(sessionId)
			if !exists {
				return
			}
			session.Email = payload.Email

			colors := [3]string{"red", "green", "blue"}
			// TODO: seeding this off time can be unsecure
			rng := rand.New(rand.NewSource(time.Now().Unix()))
			sendRandomColor := func() error {
				const stripPercentMax = 90
				const stripPercentMin = 10
				backgroundColor := colors[rng.Int()%len(colors)]
				var stripColor = colors[rng.Int()%len(colors)]
				for stripColor == backgroundColor {
					stripColor = colors[rng.Int()%len(colors)]
				}
				stripPosition := rng.Intn(stripPercentMax-stripPercentMin) + stripPercentMin

				colorMsg, err := MarshalWsMessage(WsMessage{
					Command: CommandSetBandColor,
					Payload: SetBandColorPayload{
						BackgroundColor: backgroundColor,
						StripColor:      stripColor,
						StripPosition:   uint16(stripPosition),
					}})
				if err != nil {
					return err
				}

				session.SentColorCommands = append(session.SentColorCommands, StoredColorCommand{
					BackgroundColor: backgroundColor,
					StripColor:      stripColor,
					StripPosition:   uint16(stripPosition),
					TimeStamp:       time.Now(),
				})

				err = ws.WriteMessage(websocket.TextMessage, []byte(colorMsg))
				if err != nil {
					return err
				}

				return nil
			}
			for i := 0; i < COLOR_CHALLENGES; i++ {
				time.Sleep(500 * time.Millisecond)
				err := sendRandomColor()
				if err != nil {
					log.Printf("failed to send color command: %s", err.Error())
					return
				}
			}
		case ColorCommandAck:
			payload, ok := parsed_message.Payload.(ColorCommandAckPayload)
			if !ok {
				log.Println("Invaid payload for ColorCommandAck")
				return
			}
			session, exists := app.sessionManager.GetSession(sessionId)
			if !exists {
				log.Printf("Session does not exist")
				return
			}

			// TODO: verify that timestamp is legitimate
			// need to send the color commands on an interval instead of all at
			// once
			session.SentColorCommands[payload.Index].TimeStamp = payload.Timestamp

			// Once we receive the last color, write to csv
			if payload.Index == COLOR_CHALLENGES-1 {
				folderPath := "./files/csv"
				os.MkdirAll(folderPath, os.ModePerm)
				csvFile, err := os.Create(fmt.Sprintf("%s/%s.csv", folderPath, sessionId))
				if err != nil {
					log.Println("Failed to create CSV file")
					return
				}
				defer csvFile.Close()
				writer := csv.NewWriter(csvFile)
				defer writer.Flush()

				if err := writer.Write([]string{"Background Color", "Strip Color", "Strip Position", "Timestamap"}); err != nil {
					log.Printf("Failed to write CSV header")
					return
				}

				for _, cmd := range session.SentColorCommands {
					record := []string{
						cmd.BackgroundColor,
						cmd.StripColor,
						fmt.Sprintf("%d", cmd.StripPosition),
						fmt.Sprintf("%d", cmd.TimeStamp.UnixMilli()),
					}
					if err := writer.Write(record); err != nil {
						log.Printf("Failed to write record to CSV")
						return
					}
				}
				fmt.Printf("Email: %s\n", email)

				// make http request with csv file and video file

				// send authenticationResult
				msg, err := MarshalWsMessage(WsMessage{
					Command: AuthenticationResult,
					Payload: AuthenticationResultPayload{
						Success: true,
					},
				})
				if err != nil {
					log.Printf("Failed to marshal authentication result: %s", err.Error())
					return
				}
				ws.WriteMessage(websocket.TextMessage, []byte(msg))
			}
		}
	}
}

func saveToDisk(i media.Writer, track *webrtc.TrackRemote) {
	defer i.Close()

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

func (app *App) getUserByEmbedding(embedding pgvector.Vector) (User, error) {
	var user User
	err := app.db.Select(&user, "SELECT * FROM registered_user ORDER BY embedding <-> $1 LIMIT 1", embedding)
	if err != nil {
		return User{}, err
	}
	if len(user.Embedding.Slice()) != len(embedding.Slice()) {
		return User{}, fmt.Errorf("expected embedding length %d, got %d", len(embedding.Slice()), len(user.Embedding.Slice()))
	}

	var total_distance float32 = 0
	for i := range user.Embedding.Slice() {
		diff := (user.Embedding.Slice()[i] - embedding.Slice()[i])
		total_distance += diff * diff
	}

	if total_distance > EMBEDDING_THRESHOLD {
		return User{}, fmt.Errorf("no user found")
	}

	return User{}, nil
}

type EnrollBody struct {
	Email     string    `json:"email"`
	Embedding []float32 `json:"embedding"`
}

func (app *App) enrollHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body EnrollBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: validate

	user := User{
		Email:     body.Email,
		Embedding: pgvector.NewVector(body.Embedding),
	}

	_, err = app.db.NamedExec("INSERT INTO registered_user (email, embedding) VALUES (:email, :embedding)", user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

// func writeToCSV(commands []StoredColorCommand, filename string) error {
// 	file, err := os.Create(filename)
// 	if err != nil {
// 		return fmt.Errorf("error while creating file: %w", err)
// 	}
// 	defer file.Close()
//
// 	writer := csv.NewWriter(file)
// 	defer writer.Flush()
//
// 	// Write CSV header
// 	if err := writer.Write([]string{"BackgroundColor", "StripColor", "StripPosition", "TimeStamp"}); err != nil {
// 		return fmt.Errorf("error while writing header to CSV: %w", err)
// 	}
//
// 	// Write data to CSV
// 	for _, cmd := range commands {
// 		record := []string{
// 			cmd.BackgroundColor,
// 			cmd.StripColor,
// 			fmt.Sprintf("%d", cmd.StripPosition),
// 			fmt.Sprintf("%d", cmd.TimeStamp.UnixNano()/int64(time.Millisecond)), // Convert time.Time to milliseconds since epoch
// 		}
//
// 		if err := writer.Write(record); err != nil {
// 			return fmt.Errorf("error while writing record to CSV: %w", err)
// 		}
// 	}
//
// 	return nil
// }

// func (app *App) generateColors(sessionId string) error {
// 	session, exists := app.sessionManager.GetSession(sessionId)
// 	if !exists {
// 		return fmt.Errorf("session does not exist")
// 	}
// 	ws := session.WebSocket
//
// 	// send two immediately so that the frontend always has one buffered
// 	err := sendRandomColor()
// 	if err != nil {
// 		return err
// 	}
// 	err = sendRandomColor()
// 	if err != nil {
// 		return err
// 	}
//
// 	ticker := time.NewTicker(500 * time.Millisecond)
// 	go func() {
// 		for i := 0; i < 8; i++ {
// 			select {
// 			case <-ticker.C:
// 				err = sendRandomColor()
// 				if err != nil {
// 					ticker.Stop()
// 				}
// 			}
// 		}
// 	}()
//
// 	return nil
// }
