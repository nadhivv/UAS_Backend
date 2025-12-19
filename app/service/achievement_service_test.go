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
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateAchievement_Success(t *testing.T) {
	// 1. Setup Fiber & Mocks
	app := fiber.New()
	mockAchRepo := &MockAchievementRepository{}
	mockAchRefRepo := &MockAchievementReferenceRepository{}
	mockStudentRepo := &MockStudentRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	// Inject mocks ke service
	s := NewAchievementService(mockAchRepo, mockAchRefRepo, mockStudentRepo, nil, mockUserRepo, mockRoleRepo)

	// 2. Data Dummy
	userID := uuid.New()
	roleID := uuid.New()
	studentID := uuid.New()
	mongoID := primitive.NewObjectID().Hex()

	// 3. Define Route (Simulasi Middleware Locals)
	app.Post("/achievements", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("user", &models.User{ID: userID, RoleID: roleID, FullName: "Budi"})
		return s.CreateAchievement(c)
	})

	// 4. MOCK BEHAVIOR (Urutan sesuai baris kode di Service)
	
	// s.roleRepo.GetByID(user.RoleID)
	mockRoleRepo.GetByIDFn = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{ID: roleID, Name: "Mahasiswa"}, nil
	}

	// s.studentRepo.GetByUserID(userID)
	mockStudentRepo.GetByUserIDFn = func(uID uuid.UUID) (*models.Student, error) {
		return &models.Student{ID: studentID, UserID: uID}, nil
	}

	// s.userRepo.GetByID(student.UserID)
	mockUserRepo.GetByIDFn = func(id uuid.UUID) (*models.User, error) {
		return &models.User{FullName: "Budi"}, nil
	}

	// s.achievementRepo.CreateAchievement(ctx, achievement)
	mockAchRepo.CreateAchievementFn = func(ctx context.Context, ach *models.Achievement) (string, error) {
		return mongoID, nil
	}

	// s.achievementRefRepo.CreateReference(ref)
	mockAchRefRepo.CreateReferenceFn = func(ref *models.AchievementReference) error {
		return nil
	}

	// 5. Execution
	payload := models.CreateAchievementRequest{
		Title:           "Juara 1 Hackathon",
		AchievementType: "competition",
		Points:          100,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/achievements", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, _ := app.Test(req)

	// 6. Assertions
	assert.Equal(t, 201, resp.StatusCode)
}

func TestGetAchievementByID_Success(t *testing.T) {
	app := fiber.New()
	mockAchRepo := &MockAchievementRepository{}
	mockAchRefRepo := &MockAchievementReferenceRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockStudentRepo := &MockStudentRepository{}
	mockUserRepo := &MockUserRepository{}
	
	s := NewAchievementService(mockAchRepo, mockAchRefRepo, mockStudentRepo, nil, mockUserRepo, mockRoleRepo)

	refID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	studentID := uuid.New()
	mongoID := primitive.NewObjectID().Hex()

	app.Get("/achievements/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("user", &models.User{ID: userID, RoleID: roleID})
		return s.GetAchievementByID(c)
	})

	// --- SETUP MOCKS SESUAI ALUR SERVICE ---

	// 1. s.achievementRefRepo.GetReferenceByID(refUUID)
	mockAchRefRepo.GetReferenceByIDFn = func(id uuid.UUID) (*models.AchievementReference, error) {
		return &models.AchievementReference{
			ID: refID, StudentID: studentID, MongoAchievementID: mongoID, Status: "draft",
		}, nil
	}

	// 2. s.achievementRepo.GetAchievementByID(ctx, ref.MongoAchievementID)
	mockAchRepo.GetAchievementByIDFn = func(ctx context.Context, id string) (*models.Achievement, error) {
		return &models.Achievement{
			ID: primitive.NewObjectID(), Title: "Test Ach", AchievementType: "academic",
		}, nil
	}

	// 3. s.roleRepo.GetByID(user.RoleID)
	mockRoleRepo.GetByIDFn = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{Name: "Mahasiswa"}, nil
	}

	// 4. s.studentRepo.GetByUserID(userID) (Inside Mahasiswa Case)
	mockStudentRepo.GetByUserIDFn = func(uID uuid.UUID) (*models.Student, error) {
		return &models.Student{ID: studentID}, nil
	}

	// 5. s.studentRepo.GetByID(ref.StudentID) (Step 4 Get Student Info)
	mockStudentRepo.GetByIDFn = func(id uuid.UUID) (*models.Student, error) {
		return &models.Student{ID: studentID, UserID: userID}, nil
	}

	// 6. s.userRepo.GetByID(student.UserID)
	mockUserRepo.GetByIDFn = func(id uuid.UUID) (*models.User, error) {
		return &models.User{FullName: "Budi"}, nil
	}

	// Execution
	req := httptest.NewRequest("GET", "/achievements/"+refID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestVerifyAchievement_Forbidden_NotAdvisor(t *testing.T) {
	app := fiber.New()
	mockAchRefRepo := &MockAchievementReferenceRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockLecturerRepo := &MockLecturerRepository{}
	mockStudentRepo := &MockStudentRepository{}

	s := NewAchievementService(nil, mockAchRefRepo, mockStudentRepo, mockLecturerRepo, nil, mockRoleRepo)

	refID := uuid.New()
	lecturerUserID := uuid.New()
	lecturerID := uuid.New()
	studentID := uuid.New()

	app.Post("/achievements/:id/verify", func(c *fiber.Ctx) error {
		c.Locals("user_id", lecturerUserID)
		c.Locals("user", &models.User{ID: lecturerUserID, RoleID: uuid.New()})
		return s.VerifyAchievement(c)
	})

	// 1. Reference: Status harus Submitted agar tidak 400
	mockAchRefRepo.GetReferenceByIDFn = func(id uuid.UUID) (*models.AchievementReference, error) {
		return &models.AchievementReference{ID: refID, StudentID: studentID, Status: models.AchievementStatusSubmitted}, nil
	}

	// 2. Role: Dosen Wali
	mockRoleRepo.GetByIDFn = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{Name: "Dosen Wali"}, nil
	}

	// 3. Get Lecturer Profile
	mockLecturerRepo.GetByUserIDFn = func(uID uuid.UUID) (*models.Lecturer, error) {
		return &models.Lecturer{ID: lecturerID}, nil
	}

	// 4. Get Student Profile: AdvisorID berbeda agar Forbidden
	mockStudentRepo.GetByIDFn = func(id uuid.UUID) (*models.Student, error) {
		otherLecturerID := uuid.New()
		return &models.Student{ID: studentID, AdvisorID: &otherLecturerID}, nil
	}

	req := httptest.NewRequest("POST", "/achievements/"+refID.String()+"/verify", nil)
	resp, _ := app.Test(req)

	// Harus 403 karena AdvisorID != LecturerID
	assert.Equal(t, 403, resp.StatusCode)
}