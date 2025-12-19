package service

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"UAS/app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestReportService_GetStatistics_Admin_Success(t *testing.T) {
	// 1. Setup Fiber & Mocks
	app := fiber.New()
	mockReportRepo := &MockReportRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockUserRepo := &MockUserRepository{}
	mockStudentRepo := &MockStudentRepository{}
	mockLecturerRepo := &MockLecturerRepository{}

	s := NewReportService(mockReportRepo, mockUserRepo, mockStudentRepo, mockLecturerRepo, mockRoleRepo)

	// 2. Data Dummy
	adminID := uuid.New()
	roleID := uuid.New()
	
	app.Get("/reports/statistics", func(c *fiber.Ctx) error {
		// Sesuai dengan service: currentUser, ok := c.Locals("user").(*models.User)
		c.Locals("user", &models.User{ID: adminID, RoleID: roleID})
		return s.GetStatistics(c)
	})

	// 3. MOCK BEHAVIOR
	// s.roleRepo.GetByID(currentUser.RoleID)
	mockRoleRepo.GetByIDFn = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{ID: roleID, Name: "Admin"}, nil
	}

	// s.reportRepo.GetStatistics(...)
	mockReportRepo.GetStatisticsFn = func(ctx context.Context, actorID uuid.UUID, scope string, start, end *time.Time) (*models.AchievementStats, error) {
		// Menggunakan struct kosong agar tidak error "unknown field"
		return &models.AchievementStats{}, nil
	}

	// 4. Execution
	req := httptest.NewRequest("GET", "/reports/statistics?start_date=2024-01-01&end_date=2024-12-31", nil)
	resp, _ := app.Test(req)

	// 5. Assertions
	assert.Equal(t, 200, resp.StatusCode)
	
	var resBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.True(t, resBody["success"].(bool))
}

func TestReportService_GetStatistics_Mahasiswa_Success(t *testing.T) {
	app := fiber.New()
	mockReportRepo := &MockReportRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockStudentRepo := &MockStudentRepository{}
	
	s := NewReportService(mockReportRepo, nil, mockStudentRepo, nil, mockRoleRepo)

	userID := uuid.New()
	roleID := uuid.New()
	studentID := uuid.New()

	app.Get("/reports/statistics", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{ID: userID, RoleID: roleID})
		return s.GetStatistics(c)
	})

	mockRoleRepo.GetByIDFn = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{Name: "Mahasiswa"}, nil
	}

	mockStudentRepo.GetByUserIDFn = func(uID uuid.UUID) (*models.Student, error) {
		return &models.Student{ID: studentID, UserID: uID}, nil
	}

	mockReportRepo.GetStatisticsFn = func(ctx context.Context, actorID uuid.UUID, scope string, start, end *time.Time) (*models.AchievementStats, error) {
		// Validasi bahwa service mengirim scope dan ID yang benar
		assert.Equal(t, "student", scope)
		assert.Equal(t, studentID, actorID)
		return &models.AchievementStats{}, nil
	}

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestReportService_GetStudentReport_Success_Admin(t *testing.T) {
	app := fiber.New()
	mockReportRepo := &MockReportRepository{}
	mockRoleRepo := &MockRoleRepository{}
	
	s := NewReportService(mockReportRepo, nil, nil, nil, mockRoleRepo)

	adminID := uuid.New()
	targetStudentID := uuid.New()

	app.Get("/reports/students/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{ID: adminID, RoleID: uuid.New()})
		return s.GetStudentReport(c)
	})

	// Mock Role: Admin
	mockRoleRepo.GetByIDFn = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{Name: "Admin"}, nil
	}

	mockReportRepo.GetStatisticsFn = func(ctx context.Context, actorID uuid.UUID, scope string, start, end *time.Time) (*models.AchievementStats, error) {
		return &models.AchievementStats{}, nil
	}

	req := httptest.NewRequest("GET", "/reports/students/"+targetStudentID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestReportService_GetStudentReport_Forbidden_Mahasiswa(t *testing.T) {
	app := fiber.New()
	mockRoleRepo := &MockRoleRepository{}
	mockStudentRepo := &MockStudentRepository{}
	
	s := NewReportService(nil, nil, mockStudentRepo, nil, mockRoleRepo)

	myUserID := uuid.New()
	targetStudentID := uuid.New()
	otherUserID := uuid.New()

	app.Get("/reports/students/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{ID: myUserID, RoleID: uuid.New()})
		return s.GetStudentReport(c)
	})

	mockRoleRepo.GetByIDFn = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{Name: "Mahasiswa"}, nil
	}

	// Mock Student: Mengembalikan data student milik ORANG LAIN (userID berbeda)
	mockStudentRepo.GetByIDFn = func(id uuid.UUID) (*models.Student, error) {
		return &models.Student{ID: targetStudentID, UserID: otherUserID}, nil
	}

	req := httptest.NewRequest("GET", "/reports/students/"+targetStudentID.String(), nil)
	resp, _ := app.Test(req)

	// Status 403 karena Mahasiswa tidak boleh melihat report mahasiswa lain
	assert.Equal(t, 403, resp.StatusCode)
}