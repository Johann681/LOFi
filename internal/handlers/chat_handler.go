package handlers

import (
	"context"
	"encoding/json"
	"time"

	"lofi-student-match/internal/models"
	realtime "lofi-student-match/internal/websocket"

	"github.com/gofiber/fiber/v2"
	fiberws "github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatHandler struct {
	hub      *realtime.Hub
	database *mongo.Database
}

func NewChatHandler(hub *realtime.Hub, database *mongo.Database) *ChatHandler {
	return &ChatHandler{hub: hub, database: database}
}

// Handle validates the query ID, upgrades the socket, and registers its client.
func (handler *ChatHandler) Handle(conn *fiberws.Conn) {
	studentID, err := primitive.ObjectIDFromHex(conn.Query("student_id"))
	if err != nil {
		writeSocketError(conn, "student_id must be a valid MongoDB ObjectID")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = handler.database.Collection("students").FindOne(ctx, bson.M{"_id": studentID}).Err()
	cancel()
	if err != nil {
		writeSocketError(conn, "student not found")
		return
	}

	client := realtime.NewClient(studentID, conn, handler.hub, handler.database)
	if !handler.hub.Register(client) {
		writeSocketError(conn, "student already has an active chat connection")
		return
	}
	client.Run()
}

// History returns persisted chat messages only when the requester belongs to
// the match, including conversations whose match is no longer active.
func (handler *ChatHandler) History(c *fiber.Ctx) error {
	matchID, err := primitive.ObjectIDFromHex(c.Params("match_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid match ID")
	}
	studentID, err := primitive.ObjectIDFromHex(c.Query("student_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "student_id must be a valid MongoDB ObjectID")
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
	defer cancel()
	var match models.Match
	if err := handler.database.Collection("matches").FindOne(ctx, bson.M{
		"_id":        matchID,
		"studentIds": studentID,
	}).Decode(&match); err != nil {
		if err == mongo.ErrNoDocuments {
			return fiber.NewError(fiber.StatusNotFound, "match not found for this student")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not verify match membership")
	}

	cursor, err := handler.database.Collection("messages").Find(ctx,
		bson.M{"matchId": matchID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}).SetLimit(200),
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not fetch messages")
	}
	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not decode messages")
	}
	if messages == nil {
		messages = []models.Message{}
	}
	return c.JSON(fiber.Map{"messages": messages})
}

func writeSocketError(conn *fiberws.Conn, message string) {
	payload, _ := json.Marshal(map[string]string{"type": "ERROR", "error": message})
	_ = conn.WriteMessage(fiberws.TextMessage, payload)
	_ = conn.Close()
}
