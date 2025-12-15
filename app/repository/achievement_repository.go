package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"UAS/app/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AchievementRepository interface {
	// CRUD operations
	CreateAchievement(ctx context.Context, achievement *models.Achievement) (string, error)
	GetAchievementByID(ctx context.Context, id string) (*models.Achievement, error)
	UpdateAchievement(ctx context.Context, id string, achievement *models.Achievement) error
	DeleteAchievement(ctx context.Context, id string) error
	
	// Query operations
	FindAchievements(ctx context.Context, studentIDs []string, achievementType, search string, page, limit int, sortBy, sortOrder string) ([]models.Achievement, int64, error)
	GetAchievementsByIDs(ctx context.Context, ids []string) ([]models.Achievement, error)
	
	// Attachment operations
	AddAttachment(ctx context.Context, achievementID string, attachment models.Attachment) error
	RemoveAttachment(ctx context.Context, achievementID, fileName string) error
}

type achievementRepo struct {
	Collection *mongo.Collection
}

func NewAchievementRepository(collection *mongo.Collection) AchievementRepository {
	return &achievementRepo{Collection: collection}
}

func (r *achievementRepo) CreateAchievement(ctx context.Context, achievement *models.Achievement) (string, error) {
	achievement.ID = primitive.NewObjectID()
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()
	
	result, err := r.Collection.InsertOne(ctx, achievement)
	if err != nil {
		return "", fmt.Errorf("failed to insert achievement: %w", err)
	}
	
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *achievementRepo) GetAchievementByID(ctx context.Context, id string) (*models.Achievement, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid achievement ID: %w", err)
	}
	
	var achievement models.Achievement
	err = r.Collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&achievement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find achievement: %w", err)
	}
	
	return &achievement, nil
}

func (r *achievementRepo) UpdateAchievement(ctx context.Context, id string, achievement *models.Achievement) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid achievement ID: %w", err)
	}
	
	achievement.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"title":           achievement.Title,
			"description":     achievement.Description,
			"achievementType": achievement.AchievementType,
			"details":         achievement.Details,
			"tags":            achievement.Tags,
			"points":          achievement.Points,
			"updatedAt":       achievement.UpdatedAt,
		},
	}
	
	result, err := r.Collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("failed to update achievement: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return fmt.Errorf("achievement not found")
	}
	
	return nil
}

func (r *achievementRepo) DeleteAchievement(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid achievement ID: %w", err)
	}
	
	result, err := r.Collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete achievement: %w", err)
	}
	
	if result.DeletedCount == 0 {
		return fmt.Errorf("achievement not found")
	}
	
	return nil
}

func (r *achievementRepo) FindAchievements(ctx context.Context, studentIDs []string, achievementType, search string, page, limit int, sortBy, sortOrder string) ([]models.Achievement, int64, error) {
	// Build MongoDB filter
	mongoFilter := bson.M{}
	
	if len(studentIDs) > 0 {
		mongoFilter["studentId"] = bson.M{"$in": studentIDs}
	}
	
	if achievementType != "" {
		mongoFilter["achievementType"] = achievementType
	}
	
	if search != "" {
		searchLower := strings.ToLower(search)
		mongoFilter["$or"] = []bson.M{
			{"title": bson.M{"$regex": searchLower, "$options": "i"}},
			{"description": bson.M{"$regex": searchLower, "$options": "i"}},
			{"tags": bson.M{"$in": []string{searchLower}}},
		}
	}
	
	// Count total documents
	total, err := r.Collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}
	
	// Set pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	skip := int64((page - 1) * limit)
	
	// Set sort options
	if sortBy == "" {
		sortBy = "createdAt"
	}
	if sortOrder == "" || (sortOrder != "asc" && sortOrder != "desc") {
		sortOrder = "desc"
	}
	
	sortValue := 1 // asc
	if sortOrder == "desc" {
		sortValue = -1
	}
	
	// Find options
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: sortBy, Value: sortValue}})
	
	// Execute query
	cursor, err := r.Collection.Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find achievements: %w", err)
	}
	defer cursor.Close(ctx)
	
	var achievements []models.Achievement
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, 0, fmt.Errorf("failed to decode achievements: %w", err)
	}
	
	return achievements, total, nil
}

func (r *achievementRepo) GetAchievementsByIDs(ctx context.Context, ids []string) ([]models.Achievement, error) {
	var objectIDs []primitive.ObjectID
	for _, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, fmt.Errorf("invalid achievement ID %s: %w", id, err)
		}
		objectIDs = append(objectIDs, objectID)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find achievements: %w", err)
	}
	defer cursor.Close(ctx)
	
	var achievements []models.Achievement
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, fmt.Errorf("failed to decode achievements: %w", err)
	}
	
	return achievements, nil
}

func (r *achievementRepo) AddAttachment(ctx context.Context, achievementID string, attachment models.Attachment) error {
	objectID, err := primitive.ObjectIDFromHex(achievementID)
	if err != nil {
		return fmt.Errorf("invalid achievement ID: %w", err)
	}
	
	attachment.UploadedAt = time.Now()
	update := bson.M{
		"$push": bson.M{
			"attachments": attachment,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}
	
	result, err := r.Collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("failed to add attachment: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return fmt.Errorf("achievement not found")
	}
	
	return nil
}

func (r *achievementRepo) RemoveAttachment(ctx context.Context, achievementID, fileName string) error {
	objectID, err := primitive.ObjectIDFromHex(achievementID)
	if err != nil {
		return fmt.Errorf("invalid achievement ID: %w", err)
	}
	
	update := bson.M{
		"$pull": bson.M{
			"attachments": bson.M{"fileName": fileName},
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}
	
	result, err := r.Collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("failed to remove attachment: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return fmt.Errorf("achievement not found")
	}
	
	return nil
}