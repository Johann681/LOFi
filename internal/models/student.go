package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Preferences struct {
	TargetGender string `bson:"targetGender" json:"targetGender"`
	Orientation  string `bson:"orientation" json:"orientation"`
}

type Student struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email             string             `bson:"email" json:"email"`
	PasswordHash      string             `bson:"passwordHash,omitempty" json:"-"`
	Name              string             `bson:"name" json:"name"`
	Gender            string             `bson:"gender" json:"gender"`
	Orientation       string             `bson:"orientation" json:"orientation"`
	Age               int                `bson:"age" json:"age"`
	Height            string             `bson:"height" json:"height"`
	Department        string             `bson:"department" json:"department"`
	Class             string             `bson:"class" json:"class"`
	VerificationProof string             `bson:"verificationProof" json:"verificationProof"`
	Preferences       Preferences        `bson:"preferences" json:"preferences"`
	Likes             []string           `bson:"likes" json:"likes"`
	Dislikes          []string           `bson:"dislikes" json:"dislikes"`
	CreatedAt         time.Time          `bson:"createdAt" json:"createdAt"`
}
