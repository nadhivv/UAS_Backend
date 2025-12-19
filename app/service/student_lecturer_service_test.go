package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"UAS/app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


// ===== SETUP FIBER APP untuk Testing =====
func setupStudentLecturerApp(svc *StudentLecturerService, user *models.User) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		if user != nil {
			c.Locals("user", user)
			c.Locals("user_id", user.ID)
		}
		return c.Next()
	})

	// Contoh route untuk test
	app.Get("/students", svc.GetAllStudents)
	app.Get("/students/:id", svc.GetStudentByID)
	app.Put("/students/:id/advisor", svc.UpdateStudentAdvisor)
	app.Get("/students/:id/achievements", svc.GetStudentAchievements)
	app.Get("/lecturers/:id/advisees", svc.GetLecturerAdvisees)

	return app
}

// ===== UNIT TEST CONTOH =====

// Test GetAllStudents
func TestGetAllStudents_Success(t *testing.T) {
	mockStudents := []models.Student{
		{ID: uuid.New(), StudentID: "101"},
		{ID: uuid.New(), StudentID: "102"},
	}

	svc := &StudentLecturerService{
		studentRepo: &MockStudentRepository{
			GetAllFn: func() ([]models.Student, error) {
				return mockStudents, nil
			},
		},
	}

	app := setupStudentLecturerApp(svc, nil)

	req := httptest.NewRequest("GET", "/students?page=1&limit=10", nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)

	data := respBody["data"].([]interface{})
	if len(data) != len(mockStudents) {
		t.Errorf("Expected %d students, got %d", len(mockStudents), len(data))
	}
}

// Test Update Student Advisor Assign
func TestUpdateStudentAdvisor_Assign(t *testing.T) {
	studentID := uuid.New()
	lecturerID := uuid.New()
	userID := uuid.New()

	svc := &StudentLecturerService{
		studentRepo: &MockStudentRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID, UserID: userID}, nil
			},
			UpdateAdvisorFn: func(sID uuid.UUID, aID *uuid.UUID) error {
				if sID == studentID && aID != nil && *aID == lecturerID {
					return nil
				}
				return fiber.ErrBadRequest
			},
		},
		lecturerRepo: &MockLecturerRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Lecturer, error) {
				return &models.Lecturer{ID: lecturerID}, nil
			},
		},
		userRepo: &MockUserRepository{
			GetByIDFn: func(id uuid.UUID) (*models.User, error) {
				return &models.User{FullName: "Dosen Keren"}, nil
			},
		},
	}

	app := setupStudentLecturerApp(svc, nil)

	payload := map[string]string{"advisor_id": lecturerID.String()}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/students/"+studentID.String()+"/advisor", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)

	data := respBody["data"].(map[string]interface{})
	if data["action"] != "assigned" {
		t.Errorf("Expected action 'assigned', got %s", data["action"])
	}
	advisor := data["advisor"].(map[string]interface{})
	if advisor["name"] != "Dosen Keren" {
		t.Errorf("Expected advisor name 'Dosen Keren', got %s", advisor["name"])
	}
}

// Test GetStudentAchievements Success
func TestGetStudentAchievements_Success(t *testing.T) {
	studentID := uuid.New()
	userID := uuid.New()
	mongoID := primitive.NewObjectID()

	mockUser := &models.User{ID: userID, FullName: "Mhs Berprestasi"}

	svc := &StudentLecturerService{
		studentRepo: &MockStudentRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID, UserID: userID}, nil
			},
		},
		achievementRefRepo: &MockAchievementReferenceRepository{
			GetReferencesByStudentIDFn: func(studentID uuid.UUID, status string) ([]models.AchievementReference, error) {
				return []models.AchievementReference{
					{
						ID:                 uuid.New(),
						MongoAchievementID: mongoID.Hex(),
						Status:             "verified",
						SubmittedAt:        &time.Time{},
					},
				}, nil
			},
		},
		achievementRepo: &MockAchievementRepository{
			GetAchievementByIDFn: func(ctx context.Context, id string) (*models.Achievement, error) {
				if id == mongoID.Hex() {
					return &models.Achievement{
						Title:           "Juara 1 Lomba",
						Points:          100,
						AchievementType: "Lomba",
					}, nil
				}
				return nil, nil
			},
		},
		userRepo: &MockUserRepository{
			GetByIDFn: func(id uuid.UUID) (*models.User, error) {
				return mockUser, nil
			},
		},
	}

	// Inject User via Middleware (login sebagai owner student)
	app := setupStudentLecturerApp(svc, mockUser)

	req := httptest.NewRequest("GET", "/students/"+studentID.String()+"/achievements", nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var responseBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseBody)
	data := responseBody["data"].(map[string]interface{})

	achievements := data["achievements"].([]interface{})
	if len(achievements) == 0 {
		t.Error("Should return achievements")
	}

	firstAch := achievements[0].(map[string]interface{})
	if firstAch["title"] != "Juara 1 Lomba" {
		t.Errorf("Expected title 'Juara 1 Lomba', got %s", firstAch["title"])
	}
}

// Test GetStudentByID Access Denied
func TestGetStudentByID_AccessDenied(t *testing.T) {
	ownerUserID := uuid.New()
	attackerUserID := uuid.New()
	studentID := uuid.New()

	attackerUser := &models.User{
		ID:       attackerUserID,
		RoleID:   uuid.New(),
		FullName: "Attacker",
	}

	svc := &StudentLecturerService{
		studentRepo: &MockStudentRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID, UserID: ownerUserID}, nil
			},
		},
		lecturerRepo: &MockLecturerRepository{
			GetByUserIDFn: func(uid uuid.UUID) (*models.Lecturer, error) {
				return nil, nil // attacker bukan dosen
			},
		},
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{Name: "Mahasiswa"}, nil
			},
		},
	}

	app := setupStudentLecturerApp(svc, attackerUser)

	req := httptest.NewRequest("GET", "/students/"+studentID.String(), nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 403 {
		t.Errorf("Expected 403 Forbidden, got %d", resp.StatusCode)
	}

	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	if respBody["error"] != "Access denied" {
		t.Errorf("Expected 'Access denied', got %s", respBody["error"])
	}
}
