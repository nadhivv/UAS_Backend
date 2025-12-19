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

// Helper untuk setup app dan service agar tidak ada repo yang NIL
func setupUserApp() (*fiber.App, *MockUserRepository, *MockRoleRepository, *MockStudentRepository, *MockLecturerRepository, *UserService) {
	mUser := &MockUserRepository{}
	mRole := &MockRoleRepository{}
	mStd := &MockStudentRepository{}
	mLect := &MockLecturerRepository{}
	
	svc := NewUserService(mUser, mRole, mStd, mLect)
	
	app := fiber.New()
	app.Post("/users", svc.Create)
	app.Get("/users", svc.GetAll)
	app.Get("/users/:id", svc.GetByID)
	app.Put("/users/:id", svc.Update)
	app.Delete("/users/:id", svc.Delete)
	
	return app, mUser, mRole, mStd, mLect, svc
}

func TestUserService_Create_Mahasiswa_Success(t *testing.T) {
	app, mUser, mRole, mStd, mLect, _ := setupUserApp()

	roleID := uuid.New()
	userID := uuid.New()

	// 1. Mock Role - WAJIB ada Name agar switch case di service tidak nyasar
	mRole.GetByIDFn = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{ID: roleID, Name: "Mahasiswa"}, nil
	}

	// 2. Mock User Uniqueness
	mUser.GetByUsernameFn = func(username string) (*models.User, error) { return nil, nil }
	mUser.GetByEmailFn = func(email string) (*models.User, error) { return nil, nil }

	// 3. Mock Create User
	mUser.CreateFn = func(user *models.User) (uuid.UUID, error) { return userID, nil }

	// 4. Mock Profile Mahasiswa
	// Walaupun AdvisorID nil di request, pastikan Mock Lecturer tidak bikin panic
	mLect.GetByIDFn = func(id uuid.UUID) (*models.Lecturer, error) {
		return &models.Lecturer{ID: uuid.New()}, nil
	}
	mStd.CreateFn = func(s models.Student) (uuid.UUID, error) { return uuid.New(), nil }

	// 5. Mock Fetch User Setelah Create (INI SERING JADI PENYEBAB PANIC DI BARIS 86)
	// Pastikan return data LENGKAP, jangan cuma &models.User{}
	mUser.GetByIDFn = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:       userID, 
			Username: "newuser", 
			FullName: "New User",
			RoleID:   roleID,
		}, nil
	}

	// Payload
	stdID, prodi, thn := "123456", "Informatika", "2023"
	payload := models.CreateUserRequest{
		Username:     "newuser",
		Email:        "new@example.com",
		Password:     "password123",
		FullName:     "New User",
		RoleID:       roleID.String(),
		StudentID:    &stdID,
		ProgramStudy: &prodi,
		AcademicYear: &thn,
		AdvisorID:    nil, 
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Eksekusi - Gunakan timeout agar tidak hang
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, 201, resp.StatusCode)
}

func TestUserService_GetAll_Success(t *testing.T) {
	app, mUser, _, _, _, _ := setupUserApp()

	mUser.GetAllFn = func(page, limit int) ([]models.User, int, error) {
		return []models.User{{Username: "user1"}}, 1, nil
	}

	req := httptest.NewRequest("GET", "/users?page=1&limit=10", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	app, mUser, _, _, _, _ := setupUserApp()
	userID := uuid.New()

	// Pastikan GetByIDFn di-override untuk mengembalikan nil agar memicu 404
	mUser.GetByIDFn = func(id uuid.UUID) (*models.User, error) {
		return nil, nil
	}

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 404, resp.StatusCode)
}

func TestUserService_Delete_Success(t *testing.T) {
	app, mUser, _, _, _, _ := setupUserApp()
	userID := uuid.New()

	mUser.GetByIDFn = func(id uuid.UUID) (*models.User, error) {
		return &models.User{ID: userID, IsActive: true}, nil
	}
	mUser.SoftDeleteFn = func(id uuid.UUID) error { return nil }

	req := httptest.NewRequest("DELETE", "/users/"+userID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestUserService_Update_EmailConflict(t *testing.T) {
	app, mUser, _, _, _, _ := setupUserApp()
	myID, otherID := uuid.New(), uuid.New()

	mUser.GetByIDFn = func(id uuid.UUID) (*models.User, error) {
		return &models.User{ID: myID, IsActive: true}, nil
	}
	// Mock email sudah dipakai orang lain
	mUser.GetByEmailFn = func(email string) (*models.User, error) {
		return &models.User{ID: otherID, Email: email}, nil
	}

	emailUpdate := "conflict@example.com"
	payload := models.UpdateUserRequest{Email: &emailUpdate}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("PUT", "/users/"+myID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, 409, resp.StatusCode)
}