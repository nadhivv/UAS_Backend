package service

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"UAS/app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// --- HELPER: Hash Password ---
func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes)
}

// --- SETUP FIBER APP ---
func setupAuthApp(svc *AuthService, user *models.User) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		if user != nil {
			c.Locals("user", user)
			c.Locals("user_id", user.ID)
		}
		return c.Next()
	})

	app.Post("/auth/login", svc.Login)
	app.Get("/auth/profile", svc.Profile)

	return app
}

// --- UNIT TEST LOGIN SUCCESS ---
func TestLogin_Success(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	passwordRaw := "password123"
	passwordHash := hashPassword(passwordRaw)

	mockUser := &models.User{
		ID:           userID,
		Username:     "mahasiswa1",
		Email:        "mhs@univ.ac.id",
		PasswordHash: passwordHash,
		RoleID:       roleID,
		IsActive:     true,
	}

	svc := &AuthService{
		userRepo: &MockUserRepository{
			GetByUsernameFn: func(username string) (*models.User, error) {
				if username == "mahasiswa1" {
					return mockUser, nil
				}
				return nil, nil
			},
		},
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{ID: roleID, Name: "Mahasiswa"}, nil
			},
			GetPermissionNamesByRoleIDFn: func(roleID uuid.UUID) ([]string, error) {
				return []string{"achievement.create", "achievement.read"}, nil
			},
		},
	}

	app := setupAuthApp(svc, nil)

	payload := map[string]string{
		"username": "mahasiswa1",
		"password": "password123",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)

	if respBody["token"] == nil {
		t.Error("Expected token in response")
	}
}

// --- UNIT TEST LOGIN WRONG PASSWORD ---
func TestLogin_WrongPassword(t *testing.T) {
	passwordHash := hashPassword("correct_password")

	svc := &AuthService{
		userRepo: &MockUserRepository{
			GetByUsernameFn: func(username string) (*models.User, error) {
				return &models.User{
					Username:     "mahasiswa1",
					PasswordHash: passwordHash,
					IsActive:     true,
				}, nil
			},
		},
	}

	app := setupAuthApp(svc, nil)

	payload := map[string]string{
		"username": "mahasiswa1",
		"password": "wrong_password",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != 401 {
		t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

// --- UNIT TEST PROFILE ---
func TestProfile_Student(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()

	mockUser := &models.User{
		ID:     userID,
		RoleID: roleID,
		FullName: "Budi Santoso",
	}

	svc := &AuthService{
		userRepo: &MockUserRepository{
			GetByIDFn: func(id uuid.UUID) (*models.User, error) {
				return mockUser, nil
			},
		},
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{Name: "Mahasiswa"}, nil
			},
			GetPermissionNamesByRoleIDFn: func(roleID uuid.UUID) ([]string, error) {
				return []string{}, nil
			},
		},
		studentRepo: &MockStudentRepository{
			GetByUserIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{
					ID:           uuid.New(),
					StudentID:    "0812345",
					ProgramStudy: "Informatika",
				}, nil
			},
		},
	}

	app := setupAuthApp(svc, mockUser)

	req := httptest.NewRequest("GET", "/auth/profile", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var respBody map[string]map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)

	data := respBody["data"]
	if data["studentProfile"] == nil {
		t.Error("Expected studentProfile in response data")
	}
}
