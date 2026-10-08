package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MatchID     primitive.ObjectID `bson:"matchId" json:"matchId"`
	SenderID    primitive.ObjectID `bson:"senderId" json:"senderId"`
	RecipientID primitive.ObjectID `bson:"recipientId" json:"recipientId"`
	Content     string             `bson:"content" json:"content"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}
