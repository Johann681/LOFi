package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"lofi-student-match/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const matchQueryLimit = 500

type MatchService struct {
	students *mongo.Collection
	matches  *mongo.Collection
	queue    *mongo.Collection
}

type MatchResult struct {
	Status  string             `json:"status"`
	MatchID primitive.ObjectID `json:"matchId,omitempty"`
	Student *models.Student    `json:"student,omitempty"`
}

type MatchHistoryItem struct {
	Match   models.Match    `json:"match"`
	Student *models.Student `json:"student,omitempty"`
}

type scoredStudent struct {
	student models.Student
	score   int
}

func NewMatchService(database *mongo.Database) (*MatchService, error) {
	service := &MatchService{
		students: database.Collection("students"),
		matches:  database.Collection("matches"),
		queue:    database.Collection("matching_queue"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queueIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "studentId", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "queuedAt", Value: 1}}},
	}
	if _, err := service.queue.Indexes().CreateMany(ctx, queueIndexes); err != nil {
		return nil, fmt.Errorf("create matching queue indexes: %w", err)
	}

	matchIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "studentIds", Value: 1}},
			Options: options.Index().SetUnique(true).
				SetPartialFilterExpression(bson.M{"status": models.MatchStatusActive}),
		},
		{
			Keys: bson.D{{Key: "pairKey", Value: 1}},
			Options: options.Index().SetUnique(true).
				SetPartialFilterExpression(bson.M{"pairKey": bson.M{"$type": "string"}}),
		},
		{Keys: bson.D{{Key: "studentIds", Value: 1}, {Key: "createdAt", Value: -1}}},
	}
	if _, err := service.matches.Indexes().CreateMany(ctx, matchIndexes); err != nil {
		return nil, fmt.Errorf("create match indexes: %w", err)
	}
	return service, nil
}

// FindMatch queues the requester, scores waiting candidates, and relies on a
// unique MongoDB index to atomically prevent active double-pairing.
func (service *MatchService) FindMatch(ctx context.Context, studentID primitive.ObjectID) (*MatchResult, error) {
	var requester models.Student
	if err := service.students.FindOne(ctx, bson.M{"_id": studentID}).Decode(&requester); err != nil {
		return nil, err
	}
	if result, found, err := service.activeMatchFor(ctx, studentID); err != nil {
		return nil, err
	} else if found {
		return result, nil
	}

	_, err := service.queue.UpdateOne(ctx,
		bson.M{"studentId": studentID},
		bson.M{"$setOnInsert": bson.M{"studentId": studentID, "queuedAt": time.Now().UTC()}},
		options.Update().SetUpsert(true))
	if err != nil {
		return nil, fmt.Errorf("queue student for matching: %w", err)
	}

	queueCursor, err := service.queue.Find(ctx, bson.M{"studentId": bson.M{"$ne": studentID}},
		options.Find().SetSort(bson.D{{Key: "queuedAt", Value: 1}}).SetLimit(matchQueryLimit))
	if err != nil {
		return nil, fmt.Errorf("read matching queue: %w", err)
	}
	var queueEntries []models.MatchingQueueEntry
	if err := queueCursor.All(ctx, &queueEntries); err != nil {
		return nil, fmt.Errorf("decode matching queue: %w", err)
	}
	if len(queueEntries) == 0 {
		return &MatchResult{Status: "searching"}, nil
	}

	candidateIDs := make([]primitive.ObjectID, 0, len(queueEntries))
	queueOrder := make(map[primitive.ObjectID]int, len(queueEntries))
	for index, entry := range queueEntries {
		candidateIDs = append(candidateIDs, entry.StudentID)
		queueOrder[entry.StudentID] = index
	}

	priorCursor, err := service.matches.Find(ctx, bson.M{"studentIds": studentID})
	if err != nil {
		return nil, fmt.Errorf("read previous matches: %w", err)
	}
	var priorMatches []models.Match
	if err := priorCursor.All(ctx, &priorMatches); err != nil {
		return nil, fmt.Errorf("decode previous matches: %w", err)
	}
	previouslyPaired := make(map[primitive.ObjectID]struct{}, len(priorMatches))
	for _, match := range priorMatches {
		for _, participantID := range participantIDs(match) {
			if participantID != studentID {
				previouslyPaired[participantID] = struct{}{}
			}
		}
	}

	studentCursor, err := service.students.Find(ctx, bson.M{"_id": bson.M{"$in": candidateIDs}})
	if err != nil {
		return nil, fmt.Errorf("read queued student profiles: %w", err)
	}
	var candidates []models.Student
	if err := studentCursor.All(ctx, &candidates); err != nil {
		return nil, fmt.Errorf("decode queued student profiles: %w", err)
	}
	eligible := make([]scoredStudent, 0, len(candidates))
	for _, candidate := range candidates {
		if _, paired := previouslyPaired[candidate.ID]; paired || !compatible(requester, candidate) {
			continue
		}
		eligible = append(eligible, scoredStudent{student: candidate, score: compatibilityScore(requester, candidate)})
	}
	sort.SliceStable(eligible, func(left, right int) bool {
		if eligible[left].score == eligible[right].score {
			return queueOrder[eligible[left].student.ID] < queueOrder[eligible[right].student.ID]
		}
		return eligible[left].score > eligible[right].score
	})

	for _, candidate := range eligible {
		match := models.Match{
			ID:         primitive.NewObjectID(),
			Student1ID: requester.ID,
			Student2ID: candidate.student.ID,
			StudentIDs: []primitive.ObjectID{requester.ID, candidate.student.ID},
			PairKey:    pairKey(requester.ID, candidate.student.ID),
			Status:     models.MatchStatusActive,
			CreatedAt:  time.Now().UTC(),
		}
		if _, err := service.matches.InsertOne(ctx, match); err != nil {
			if mongo.IsDuplicateKeyError(err) {
				if result, found, lookupErr := service.activeMatchFor(ctx, studentID); lookupErr != nil {
					return nil, lookupErr
				} else if found {
					return result, nil
				}
				continue
			}
			return nil, fmt.Errorf("create match: %w", err)
		}
		// Mongo's active-match unique index remains authoritative if cleanup fails.
		_, _ = service.queue.DeleteMany(ctx, bson.M{"studentId": bson.M{"$in": match.StudentIDs}})
		return &MatchResult{Status: "matched", MatchID: match.ID, Student: &candidate.student}, nil
	}
	return &MatchResult{Status: "searching"}, nil
}

func (service *MatchService) GetHistory(ctx context.Context, studentID primitive.ObjectID) ([]MatchHistoryItem, error) {
	if err := service.students.FindOne(ctx, bson.M{"_id": studentID}).Err(); err != nil {
		return nil, err
	}
	cursor, err := service.matches.Find(ctx, bson.M{"studentIds": studentID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("read match history: %w", err)
	}
	var matches []models.Match
	if err := cursor.All(ctx, &matches); err != nil {
		return nil, fmt.Errorf("decode match history: %w", err)
	}
	if len(matches) == 0 {
		return []MatchHistoryItem{}, nil
	}

	otherIDs := make([]primitive.ObjectID, 0, len(matches))
	for _, match := range matches {
		for _, id := range participantIDs(match) {
			if id != studentID {
				otherIDs = append(otherIDs, id)
			}
		}
	}
	studentCursor, err := service.students.Find(ctx, bson.M{"_id": bson.M{"$in": otherIDs}})
	if err != nil {
		return nil, fmt.Errorf("read matched student profiles: %w", err)
	}
	var students []models.Student
	if err := studentCursor.All(ctx, &students); err != nil {
		return nil, fmt.Errorf("decode matched student profiles: %w", err)
	}
	studentsByID := make(map[primitive.ObjectID]*models.Student, len(students))
	for index := range students {
		studentsByID[students[index].ID] = &students[index]
	}

	history := make([]MatchHistoryItem, 0, len(matches))
	for _, match := range matches {
		otherID := match.Student1ID
		if otherID == studentID {
			otherID = match.Student2ID
		}
		if match.Student1ID.IsZero() {
			for _, id := range participantIDs(match) {
				if id != studentID {
					otherID = id
					break
				}
			}
		}
		history = append(history, MatchHistoryItem{Match: match, Student: studentsByID[otherID]})
	}
	return history, nil
}

func (service *MatchService) activeMatchFor(ctx context.Context, studentID primitive.ObjectID) (*MatchResult, bool, error) {
	var match models.Match
	err := service.matches.FindOne(ctx, bson.M{"studentIds": studentID, "status": models.MatchStatusActive}).Decode(&match)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("find active match: %w", err)
	}
	partnerID := match.Student1ID
	if partnerID == studentID {
		partnerID = match.Student2ID
	}
	var partner models.Student
	if err := service.students.FindOne(ctx, bson.M{"_id": partnerID}).Decode(&partner); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &MatchResult{Status: "matched", MatchID: match.ID}, true, nil
		}
		return nil, false, fmt.Errorf("find matched student: %w", err)
	}
	return &MatchResult{Status: "matched", MatchID: match.ID, Student: &partner}, true, nil
}

func compatible(first, second models.Student) bool {
	return preferenceMatches(first.Preferences.TargetGender, second.Gender) &&
		preferenceMatches(second.Preferences.TargetGender, first.Gender) &&
		preferenceMatches(first.Preferences.Orientation, second.Orientation) &&
		preferenceMatches(second.Preferences.Orientation, first.Orientation)
}

func preferenceMatches(preference, value string) bool {
	preference, value = normalize(preference), normalize(value)
	return preference == "" || preference == "any" || preference == value
}

func compatibilityScore(first, second models.Student) int {
	score := 3*overlapCount(first.Likes, second.Likes) + overlapCount(first.Dislikes, second.Dislikes)
	if normalize(first.Department) == normalize(second.Department) {
		score += 2
	}
	if normalize(first.Class) == normalize(second.Class) {
		score++
	}
	return score - 2*(overlapCount(first.Likes, second.Dislikes)+overlapCount(first.Dislikes, second.Likes))
}

func overlapCount(first, second []string) int {
	firstSet := make(map[string]struct{}, len(first))
	for _, item := range first {
		if item = normalize(item); item != "" {
			firstSet[item] = struct{}{}
		}
	}
	seen := make(map[string]struct{})
	count := 0
	for _, item := range second {
		item = normalize(item)
		if _, exists := firstSet[item]; !exists || item == "" {
			continue
		}
		if _, counted := seen[item]; !counted {
			seen[item] = struct{}{}
			count++
		}
	}
	return count
}

func normalize(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func pairKey(first, second primitive.ObjectID) string {
	firstHex, secondHex := first.Hex(), second.Hex()
	if firstHex > secondHex {
		firstHex, secondHex = secondHex, firstHex
	}
	return firstHex + ":" + secondHex
}

func participantIDs(match models.Match) []primitive.ObjectID {
	if len(match.StudentIDs) != 0 {
		return match.StudentIDs
	}
	return []primitive.ObjectID{match.Student1ID, match.Student2ID}
}
