package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	MatchStatusActive    = "active"
	MatchStatusUnmatched = "unmatched"
)

type Match struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Student1ID primitive.ObjectID   `bson:"student1Id" json:"student1Id"`
	Student2ID primitive.ObjectID   `bson:"student2Id" json:"student2Id"`
	StudentIDs []primitive.ObjectID `bson:"studentIds" json:"-"`
	PairKey    string               `bson:"pairKey" json:"-"`
	Status     string               `bson:"status" json:"status"`
	CreatedAt  time.Time            `bson:"createdAt" json:"createdAt"`
}

type MatchingQueueEntry struct {
	StudentID primitive.ObjectID `bson:"studentId" json:"studentId"`
	QueuedAt  time.Time          `bson:"queuedAt" json:"queuedAt"`
}
