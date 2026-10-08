package websocket

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	fiberws "github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"lofi-student-match/internal/models"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 45 * time.Second
	messageLimit   = 4000
	outboundBuffer = 64
)

type Client struct {
	StudentID primitive.ObjectID
	conn      *fiberws.Conn
	send      chan []byte
	hub       *Hub
	matches   *mongo.Collection
	messages  *mongo.Collection
}

type incomingMessage struct {
	MatchID string `json:"matchId"`
	Content string `json:"content"`
}

type chatEvent struct {
	Type    string         `json:"type"`
	Message models.Message `json:"message"`
}

type errorEvent struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

func NewClient(studentID primitive.ObjectID, conn *fiberws.Conn, hub *Hub, database *mongo.Database) *Client {
	return &Client{
		StudentID: studentID,
		conn:      conn,
		send:      make(chan []byte, outboundBuffer),
		hub:       hub,
		matches:   database.Collection("matches"),
		messages:  database.Collection("messages"),
	}
}

// Run starts the single writer and owns the read loop on the current goroutine.
func (client *Client) Run() {
	go client.writePump()
	client.readPump()
}

func (client *Client) readPump() {
	defer client.hub.Unregister(client)
	defer client.conn.Close()

	client.conn.SetReadLimit(64 * 1024)
	_ = client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, payload, err := client.conn.ReadMessage()
		if err != nil {
			return
		}
		client.handleMessage(payload)
	}
}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer client.conn.Close()

	for {
		select {
		case payload, open := <-client.send:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !open {
				_ = client.conn.WriteMessage(fiberws.CloseMessage, []byte{})
				return
			}
			if err := client.conn.WriteMessage(fiberws.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(fiberws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) handleMessage(payload []byte) {
	var request incomingMessage
	if err := json.Unmarshal(payload, &request); err != nil {
		client.sendError("invalid message JSON")
		return
	}
	request.Content = strings.TrimSpace(request.Content)
	if request.Content == "" || len(request.Content) > messageLimit {
		client.sendError("content must be between 1 and 4000 bytes")
		return
	}
	matchID, err := primitive.ObjectIDFromHex(request.MatchID)
	if err != nil {
		client.sendError("matchId must be a valid MongoDB ObjectID")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var match models.Match
	err = client.matches.FindOne(ctx, bson.M{
		"_id":        matchID,
		"studentIds": client.StudentID,
		"status":     models.MatchStatusActive,
	}).Decode(&match)
	if err != nil {
		client.sendError("active match not found for this student")
		return
	}

	recipientID := match.Student1ID
	if recipientID == client.StudentID {
		recipientID = match.Student2ID
	}
	if recipientID.IsZero() {
		client.sendError("match has no valid recipient")
		return
	}

	message := models.Message{
		ID:          primitive.NewObjectID(),
		MatchID:     matchID,
		SenderID:    client.StudentID,
		RecipientID: recipientID,
		Content:     request.Content,
		CreatedAt:   time.Now().UTC(),
	}
	go client.persistMessage(message)

	event, err := json.Marshal(chatEvent{Type: "CHAT_MESSAGE", Message: message})
	if err != nil {
		client.sendError("could not encode chat message")
		return
	}
	client.hub.RouteMessage(recipientID.Hex(), event)
	client.hub.RouteMessage(client.StudentID.Hex(), event)
}

func (client *Client) persistMessage(message models.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.messages.InsertOne(ctx, message); err != nil {
		log.Printf("persist chat message %s failed: %v", message.ID.Hex(), err)
	}
}

func (client *Client) sendError(message string) {
	payload, err := json.Marshal(errorEvent{Type: "ERROR", Error: message})
	if err != nil {
		return
	}
	select {
	case client.send <- payload:
	default:
	}
}
