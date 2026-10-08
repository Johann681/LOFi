package handlers

import (
	"context"
	"errors"

	"lofi-student-match/internal/services"
	realtime "lofi-student-match/internal/websocket"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MatchHandler struct {
	service *services.MatchService
	hub     *realtime.Hub
}

type findMatchRequest struct {
	StudentID string `json:"studentId"`
}

func NewMatchHandler(service *services.MatchService, hubs ...*realtime.Hub) *MatchHandler {
	handler := &MatchHandler{service: service}
	if len(hubs) != 0 {
		handler.hub = hubs[0]
	}
	return handler
}

// Find accepts a student's ObjectID and attempts an immediate queue match.
func (handler *MatchHandler) Find(c *fiber.Ctx) error {
	var request findMatchRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON request body")
	}
	studentID, err := primitive.ObjectIDFromHex(request.StudentID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "studentId must be a valid MongoDB ObjectID")
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), databaseOperationTimeout)
	defer cancel()
	result, err := handler.service.FindMatch(ctx, studentID)
	if err != nil {
		return matchServiceError(err)
	}
	if result.Status == "matched" && handler.hub != nil && result.Student != nil {
		handler.hub.NotifyMatchFound(studentID.Hex(), result.Student.ID.Hex(), fiber.Map{
			"matchId": result.MatchID.Hex(),
			"status":  result.Status,
		})
	}
	status := fiber.StatusOK
	if result.Status == "searching" {
		status = fiber.StatusAccepted
	}
	return c.Status(status).JSON(result)
}

// History returns each prior/current match and the other student's profile.
func (handler *MatchHandler) History(c *fiber.Ctx) error {
	studentID, err := primitive.ObjectIDFromHex(c.Params("student_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "student_id must be a valid MongoDB ObjectID")
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), databaseOperationTimeout)
	defer cancel()
	history, err := handler.service.GetHistory(ctx, studentID)
	if err != nil {
		return matchServiceError(err)
	}
	return c.JSON(fiber.Map{"matches": history})
}

func matchServiceError(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fiber.NewError(fiber.StatusNotFound, "student not found")
	}
	return fiber.NewError(fiber.StatusInternalServerError, "matchmaking request failed")
}
