package service

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"UAS/app/models"
	"UAS/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_Login_Success(t *testing.T) {
	t.Parallel() // Mengizinkan test ini berjalan bersamaan dengan yang lain

	// 1. Setup Lokal (Isolasi memori)
	app := fiber.New()
	mUser := &MockUserRepository{}
	mRole := &MockRoleRepository{}
	mStd := &MockStudentRepository{}
	mLect := &MockLecturerRepository{}
	s := NewAuthService(mUser, mRole, mStd, mLect)

	// 2. Data Dummy
	roleID := uuid.New()
	userID := uuid.New()
	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)
	
	user := &models.User{
		ID:           userID,
		Username:     "user_sukses",
		Email:        "sukses@test.com",
		PasswordHash: hashedPassword,
		IsActive:     true,
		RoleID:       roleID,
	}

	app.Post("/auth/login", s.Login)

	// 3. Setup Mock Behavior khusus case ini
	mUser.GetByUsernameFn = func(username string) (*models.User, error) {
		return user, nil
	}
	mUser.GetByEmailFn = func(email string) (*models.User, error) {
		return nil, nil
	}
	mRole.GetByIDFn = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{ID: roleID, Name: "Mahasiswa"}, nil
	}
	mRole.GetPermissionNamesByRoleIDFn = func(id uuid.UUID) ([]string, error) {
		return []string{"read"}, nil
	}

	// 4. Execution
	payload := models.LoginRequest{Username: "user_sukses", Password: password}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, _ := app.Test(req)

	// 5. Assertions
	assert.Equal(t, 200, resp.StatusCode)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	t.Parallel()

	// 1. Setup Lokal
	app := fiber.New()
	mUser := &MockUserRepository{}
	mRole := &MockRoleRepository{} // Walau tidak terpanggil, harus diinject agar tidak nil
	mStd := &MockStudentRepository{}
	mLect := &MockLecturerRepository{}
	s := NewAuthService(mUser, mRole, mStd, mLect)

	hashedPassword, _ := utils.HashPassword("password_benar")
	user := &models.User{
		Username:     "user_test",
		PasswordHash: hashedPassword,
		IsActive:     true,
	}

	app.Post("/auth/login", s.Login)

	// 2. Setup Mock Behavior
	mUser.GetByUsernameFn = func(username string) (*models.User, error) {
		return user, nil
	}
	mUser.GetByEmailFn = func(email string) (*models.User, error) {
		return nil, nil
	}

	// 3. Execution (Password Salah)
	payload := models.LoginRequest{Username: "user_test", Password: "password_salah"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, _ := app.Test(req)

	// 4. Assertions
	assert.Equal(t, 401, resp.StatusCode)
}

func TestAuthService_Profile_Mahasiswa(t *testing.T) {
	t.Parallel()

	// 1. Setup Lokal
	app := fiber.New()
	mUser := &MockUserRepository{}
	mRole := &MockRoleRepository{}
	mStd := &MockStudentRepository{}
	mLect := &MockLecturerRepository{}
	s := NewAuthService(mUser, mRole, mStd, mLect)

	userID := uuid.New()
	roleID := uuid.New()

	app.Get("/auth/profile", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return s.Profile(c)
	})

	// 2. Setup Mock Behavior
	mUser.GetByIDFn = func(id uuid.UUID) (*models.User, error) {
		return &models.User{ID: userID, RoleID: roleID, FullName: "Budi"}, nil
	}
	mRole.GetByIDFn = func(id uuid.UUID) (*models.Role, error) {
		return &models.Role{ID: roleID, Name: "Mahasiswa"}, nil
	}
	mRole.GetPermissionNamesByRoleIDFn = func(id uuid.UUID) ([]string, error) {
		return []string{"read"}, nil
	}
	mStd.GetByUserIDFn = func(uID uuid.UUID) (*models.Student, error) {
		return &models.Student{UserID: uID, StudentID: "123"}, nil
	}

	// 3. Execution
	req := httptest.NewRequest("GET", "/auth/profile", nil)
	resp, _ := app.Test(req)

	// 4. Assertions
	assert.Equal(t, 200, resp.StatusCode)
}

func TestAuthService_ChangePassword_WrongCurrentPassword(t *testing.T) {
	t.Parallel()

	// 1. Setup Lokal
	app := fiber.New()
	mUser := &MockUserRepository{}
	mRole := &MockRoleRepository{}
	mStd := &MockStudentRepository{}
	mLect := &MockLecturerRepository{}
	s := NewAuthService(mUser, mRole, mStd, mLect)

	userID := uuid.New()
	hashedOldPassword, _ := utils.HashPassword("old_pass")

	app.Post("/auth/change-password", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return s.ChangePassword(c)
	})

	// 2. Setup Mock Behavior
	mUser.GetByIDFn = func(id uuid.UUID) (*models.User, error) {
		return &models.User{ID: userID, PasswordHash: hashedOldPassword}, nil
	}

	// 3. Execution
	payload := map[string]string{
		"currentPassword": "salah_password",
		"newPassword":     "new_pass123",
		"confirmPassword": "new_pass123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, _ := app.Test(req)

	// 4. Assertions
	assert.Equal(t, 400, resp.StatusCode)
}