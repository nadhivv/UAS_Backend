package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"UAS/app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ====== CONTOH TEST ======

func setupTestApp(svc *AchievementService, user *models.User) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", user)
		c.Locals("user_id", user.ID)
		return c.Next()
	})

	app.Post("/achievements", svc.CreateAchievement)
	app.Get("/achievements/:id", svc.GetAchievementByID)
	app.Put("/achievements/:id", svc.UpdateAchievement)
	app.Post("/submit/:id", svc.SubmitAchievement)

	return app
}

func TestCreateAchievement_Success(t *testing.T) {
	userID := uuid.New()

	svc := &AchievementService{
		achievementRepo:    &MockAchievementRepository{
			CreateAchievementFn: func(ctx context.Context, achievement *models.Achievement) (string, error) {
				return uuid.New().String(), nil
			},
		},
		achievementRefRepo: &MockAchievementReferenceRepository{
			CreateReferenceFn: func(ref *models.AchievementReference) error {
				return nil
			},
		},
	}

	app := setupTestApp(svc, &models.User{
		ID: userID,
	})

	payload := map[string]interface{}{
		"title":            "Juara 1 Hackathon",
		"achievement_type": "competition",
		"description":      "Menang lomba",
		"points":           100,
	}

	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(
		"POST",
		"/achievements",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}
