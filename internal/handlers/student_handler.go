package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"lofi-student-match/internal/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const databaseOperationTimeout = 5 * time.Second

type StudentHandler struct {
	students *mongo.Collection
}

func NewStudentHandler(database *mongo.Database) *StudentHandler {
	return &StudentHandler{students: database.Collection("students")}
}

type registerStudentRequestPreferences struct {
	TargetGender string `json:"targetGender"`
	Orientation  string `json:"orientation"`
}

type registerStudentRequest struct {
	Name              string                            `json:"name"`
	Email             string                            `json:"email"`
	Password          string                            `json:"password"`
	Gender            string                            `json:"gender"`
	Orientation       string                            `json:"orientation"`
	Age               int                               `json:"age"`
	Height            string                            `json:"height"`
	Department        string                            `json:"department"`
	Class             string                            `json:"class"`
	VerificationProof string                            `json:"verificationProof"`
	Preferences       registerStudentRequestPreferences `json:"preferences"`
	Likes             []string                          `json:"likes"`
	Dislikes          []string                          `json:"dislikes"`
}

type loginStudentRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register validates the submitted profile, assigns server-owned fields, and
// inserts the new student into MongoDB.
func (handler *StudentHandler) Register(c *fiber.Ctx) error {
	var request registerStudentRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON request body")
	}

	if err := validateRegistration(request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	email := strings.ToLower(strings.TrimSpace(request.Email))
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not secure password")
	}

	ctx, cancel := context.WithTimeout(context.Background(), databaseOperationTimeout)
	defer cancel()

	var existing models.Student
	if err := handler.students.FindOne(ctx, bson.M{"email": email}).Decode(&existing); err == nil {
		return fiber.NewError(fiber.StatusConflict, "account already exists for this email")
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return fiber.NewError(fiber.StatusInternalServerError, "could not check email availability")
	}

	student := models.Student{
		ID:                primitive.NewObjectID(),
		Email:             email,
		PasswordHash:      string(passwordHash),
		Name:              strings.TrimSpace(request.Name),
		Gender:            strings.TrimSpace(request.Gender),
		Orientation:       strings.TrimSpace(request.Orientation),
		Age:               request.Age,
		Height:            strings.TrimSpace(request.Height),
		Department:        strings.TrimSpace(request.Department),
		Class:             strings.TrimSpace(request.Class),
		VerificationProof: strings.TrimSpace(request.VerificationProof),
		Preferences: models.Preferences{
			TargetGender: strings.TrimSpace(request.Preferences.TargetGender),
			Orientation:  strings.TrimSpace(request.Preferences.Orientation),
		},
		Likes:     request.Likes,
		Dislikes:  request.Dislikes,
		CreatedAt: time.Now().UTC(),
	}

	if _, err := handler.students.InsertOne(ctx, student); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not save student")
	}

	student.PasswordHash = ""
	return c.Status(fiber.StatusCreated).JSON(student)
}

func (handler *StudentHandler) Login(c *fiber.Ctx) error {
	var request loginStudentRequest
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON request body")
	}

	email := strings.ToLower(strings.TrimSpace(request.Email))
	password := strings.TrimSpace(request.Password)
	if email == "" || password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), databaseOperationTimeout)
	defer cancel()

	var student models.Student
	if err := handler.students.FindOne(ctx, bson.M{"email": email}).Decode(&student); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not log in")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(student.PasswordHash), []byte(password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")
	}

	student.PasswordHash = ""
	return c.JSON(student)
}

// GetByID parses a MongoDB ObjectID and returns the matching profile.
func (handler *StudentHandler) GetByID(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid student ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), databaseOperationTimeout)
	defer cancel()

	var student models.Student
	if err := handler.students.FindOne(ctx, bson.M{"_id": id}).Decode(&student); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fiber.NewError(fiber.StatusNotFound, "student not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not fetch student")
	}

	return c.JSON(student)
}

func validateRegistration(request registerStudentRequest) error {
	if strings.TrimSpace(request.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(request.Email) == "" {
		return errors.New("email is required")
	}
	if !strings.Contains(strings.TrimSpace(request.Email), "@") || !strings.Contains(strings.TrimSpace(request.Email), ".") {
		return errors.New("email must be a valid student email")
	}
	if strings.TrimSpace(request.Password) == "" {
		return errors.New("password is required")
	}
	if len(strings.TrimSpace(request.Password)) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	if strings.TrimSpace(request.Gender) == "" {
		return errors.New("gender is required")
	}
	if strings.TrimSpace(request.Orientation) == "" {
		return errors.New("orientation is required")
	}
	if request.Age <= 0 {
		return errors.New("age must be greater than zero")
	}
	if strings.TrimSpace(request.Height) == "" {
		return errors.New("height is required")
	}
	if strings.TrimSpace(request.Department) == "" {
		return errors.New("department is required")
	}
	if strings.TrimSpace(request.Class) == "" {
		return errors.New("class is required")
	}
	if strings.TrimSpace(request.VerificationProof) == "" {
		return errors.New("verificationProof is required")
	}
	if strings.TrimSpace(request.Preferences.TargetGender) == "" {
		return errors.New("preferences.targetGender is required")
	}
	if strings.TrimSpace(request.Preferences.Orientation) == "" {
		return errors.New("preferences.orientation is required")
	}
	return nil
}
