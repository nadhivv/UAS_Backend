package service

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"UAS/app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStudentLecturerService_GetAllStudents_Success(t *testing.T) {
	// 1. Setup Fiber & Mocks
	app := fiber.New()
	mockStudentRepo := &MockStudentRepository{}
	mockLecturerRepo := &MockLecturerRepository{}
	mockUserRepo := &MockUserRepository{}
	
	s := NewStudentLecturerService(mockStudentRepo, mockLecturerRepo, mockUserRepo, nil, nil, nil)

	// 2. Data Dummy
	studentUserID := uuid.New()
	advisorUserID := uuid.New()
	studentID := uuid.New()
	advisorID := uuid.New()

	app.Get("/students", s.GetAllStudents)

	// 3. MOCK BEHAVIOR
	// s.studentRepo.GetAll()
	mockStudentRepo.GetAllFn = func() ([]models.Student, error) {
		return []models.Student{
			{ID: studentID, UserID: studentUserID, StudentID: "12345", AdvisorID: &advisorID},
		}, nil
	}

	// s.userRepo.GetByID(student.UserID)
	mockUserRepo.GetByIDFn = func(id uuid.UUID) (*models.User, error) {
		if id == studentUserID {
			return &models.User{ID: studentUserID, FullName: "Mahasiswa Test", Email: "std@mail.com"}, nil
		}
		if id == advisorUserID {
			return &models.User{ID: advisorUserID, FullName: "Dosen Pembimbing"}, nil
		}
		return nil, nil
	}

	// s.lecturerRepo.GetByID(*student.AdvisorID)
	mockLecturerRepo.GetByIDFn = func(id uuid.UUID) (*models.Lecturer, error) {
		return &models.Lecturer{ID: advisorID, UserID: advisorUserID}, nil
	}

	// 4. Execution
	req := httptest.NewRequest("GET", "/students", nil)
	resp, _ := app.Test(req)

	// 5. Assertions
	assert.Equal(t, 200, resp.StatusCode)
	
	var resBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.True(t, resBody["success"].(bool))
	assert.Equal(t, 1, int(resBody["total"].(float64)))
}

func TestStudentLecturerService_GetStudentByID_NotFound(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := &MockStudentRepository{}
	s := NewStudentLecturerService(mockStudentRepo, nil, nil, nil, nil, nil)

	randomID := uuid.New()
	app.Get("/students/:id", s.GetStudentByID)

	// s.studentRepo.GetByID(studentID) -> return nil (Not Found)
	mockStudentRepo.GetByIDFn = func(id uuid.UUID) (*models.Student, error) {
		return nil, nil
	}

	req := httptest.NewRequest("GET", "/students/"+randomID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 404, resp.StatusCode)
}

func TestStudentLecturerService_UpdateStudentAdvisor_Success(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := &MockStudentRepository{}
	mockLecturerRepo := &MockLecturerRepository{}
	mockUserRepo := &MockUserRepository{}
	
	s := NewStudentLecturerService(mockStudentRepo, mockLecturerRepo, mockUserRepo, nil, nil, nil)

	studentID := uuid.New()
	advisorID := uuid.New()
	studentUserID := uuid.New()
	advisorUserID := uuid.New()

	app.Put("/students/:id/advisor", s.UpdateStudentAdvisor)

	// 1. Check if student exists
	mockStudentRepo.GetByIDFn = func(id uuid.UUID) (*models.Student, error) {
		return &models.Student{ID: studentID, UserID: studentUserID, StudentID: "123"}, nil
	}

	// 2. Check if advisor exists
	mockLecturerRepo.GetByIDFn = func(id uuid.UUID) (*models.Lecturer, error) {
		return &models.Lecturer{ID: advisorID, UserID: advisorUserID}, nil
	}

	// 3. Get User Details for response
	mockUserRepo.GetByIDFn = func(id uuid.UUID) (*models.User, error) {
		if id == advisorUserID {
			return &models.User{FullName: "Dosen Baru"}, nil
		}
		return &models.User{FullName: "Mahasiswa Test"}, nil
	}

	// 4. Mock Update Action
	mockStudentRepo.UpdateAdvisorFn = func(sid uuid.UUID, aid *uuid.UUID) error {
		return nil
	}

	// Payload
	payload := map[string]string{"advisor_id": advisorID.String()}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/students/"+studentID.String()+"/advisor", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestStudentLecturerService_GetLecturerAdvisees_Success(t *testing.T) {
	app := fiber.New()
	mockLecturerRepo := &MockLecturerRepository{}
	mockUserRepo := &MockUserRepository{}
	mockStudentRepo := &MockStudentRepository{}
	mockAchRefRepo := &MockAchievementReferenceRepository{}

	s := NewStudentLecturerService(mockStudentRepo, mockLecturerRepo, mockUserRepo, nil, nil, mockAchRefRepo)

	lecturerID := uuid.New()
	lecturerUserID := uuid.New()
	studentID := uuid.New()
	studentUserID := uuid.New()

	app.Get("/lecturers/:id/advisees", s.GetLecturerAdvisees)

	// 1. Mock Lecturer
	mockLecturerRepo.GetByIDFn = func(id uuid.UUID) (*models.Lecturer, error) {
		return &models.Lecturer{ID: lecturerID, UserID: lecturerUserID}, nil
	}

	// 2. Mock Advisees List
	mockLecturerRepo.GetAdviseesFn = func(id uuid.UUID, page, limit int) ([]models.Student, int, error) {
		return []models.Student{
			{ID: studentID, UserID: studentUserID, StudentID: "555"},
		}, 1, nil
	}

	// 3. Mock User Details
	mockUserRepo.GetByIDFn = func(id uuid.UUID) (*models.User, error) {
		return &models.User{FullName: "Test Name"}, nil
	}

	// 4. Mock Achievement Stats (loop inside service)
	mockAchRefRepo.GetReferencesByStudentIDFn = func(id uuid.UUID, status string) ([]models.AchievementReference, error) {
		return []models.AchievementReference{
			{Status: "verified"},
			{Status: "draft"},
		}, nil
	}

	req := httptest.NewRequest("GET", "/lecturers/"+lecturerID.String()+"/advisees", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
	
	var resBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&resBody)
	data := resBody["data"].(map[string]interface{})
	advisees := data["advisees"].([]interface{})
	
	// Cek stats verified (harus 1 dari dummy di atas)
	firstAdvisee := advisees[0].(map[string]interface{})
	stats := firstAdvisee["achievement_stats"].(map[string]interface{})
	assert.Equal(t, 1.0, stats["verified"])
}