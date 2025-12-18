package service

import (
	"errors"
	"fmt"
	"time"

	"UAS/app/models"
	"UAS/app/repository"
	"UAS/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo     repository.UserRepository
	roleRepo     repository.RoleRepository
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

// GetAll godoc
// @Summary Get all users
// @Description Get list of all active users with pagination. Admin only.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Items per page" minimum(1) maximum(100) default(10)
// @Success 200 {object} map[string]interface{} "List of users with pagination"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users [get]
func (s *UserService) GetAll(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get active users with pagination
	users, total, err := s.userRepo.GetAll(page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get users",
			"details": err.Error(),
		})
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit // ceil division
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"data": users,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}

// GetInactiveUsers godoc
// @Summary Get inactive users
// @Description Get list of inactive/soft-deleted users. Admin only.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Items per page" minimum(1) maximum(100) default(10)
// @Success 200 {object} map[string]interface{} "List of inactive users"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/inactive [get]
func (s *UserService) GetInactiveUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get inactive users
	users, total, err := s.userRepo.GetInactiveUsers(page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get inactive users",
			"details": err.Error(),
		})
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"data": users,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}

// GetByID godoc
// @Summary Get user by ID
// @Description Get user details by ID. Admin only.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{} "User details"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid user ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 404 {object} map[string]interface{} "Not Found - User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id} [get]
func (s *UserService) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Get active user only
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get user",
			"details": err.Error(),
		})
	}
	if user == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found or inactive",
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
}

// Create godoc
// @Summary Create new user
// @Description Create new user with role-specific profile (Mahasiswa/Dosen Wali). Admin only.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateUserRequest true "User data"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid data or missing required fields"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 409 {object} map[string]interface{} "Conflict - Username/email already exists"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users [post]
func (s *UserService) Create(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to validate role",
			"details": err.Error(),
		})
	}
	if role == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	existingUser, _ := s.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}

	existingUser, _ = s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Email already exists",
		})
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to hash password",
			"details": err.Error(),
		})
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		RoleID:       roleID,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	userID, err := s.userRepo.Create(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create user",
			"details": err.Error(),
		})
	}
	user.ID = userID

	if err := s.createUserProfile(user, &req, role.Name); err != nil {
		if deleteErr := s.userRepo.HardDelete(user.ID); deleteErr != nil {
			fmt.Printf("CRITICAL: Failed to rollback user creation: %v\n", deleteErr)
		}
		
		return c.Status(400).JSON(fiber.Map{
			"error": "Failed to create user profile",
			"details": err.Error(), 
		})
	}

	createdUser, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch created user",
			"details": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "User created successfully",
		"data":    createdUser,
	})
}

func (s *UserService) createUserProfile(user *models.User, req *models.CreateUserRequest, roleName string) error {
	switch roleName {
	case "Mahasiswa":
		if req.StudentID == nil || *req.StudentID == "" {
			return errors.New("studentId is required for Mahasiswa role")
		}
		if req.ProgramStudy == nil || *req.ProgramStudy == "" {
			return errors.New("programStudy is required for Mahasiswa role")
		}
		if req.AcademicYear == nil || *req.AcademicYear == "" {
			return errors.New("academicYear is required for Mahasiswa role")
		}

		student := models.Student{
			ID:           uuid.New(),  
			UserID:       user.ID,
			StudentID:    *req.StudentID,
			ProgramStudy: *req.ProgramStudy,
			AcademicYear: *req.AcademicYear,
			AdvisorID:    nil,
			CreatedAt:    time.Now(),
		}

		// Handle advisor jika ada
		if req.AdvisorID != nil && *req.AdvisorID != "" {
			advisorID, err := uuid.Parse(*req.AdvisorID)
			if err != nil {
				return fmt.Errorf("invalid advisor ID format: %w", err)
			}
			
			// Validasi advisor exist
			advisor, err := s.lecturerRepo.GetByID(advisorID)
			if err != nil {
				return fmt.Errorf("error checking advisor: %w", err)
			}
			if advisor == nil {
				return fmt.Errorf("advisor not found with ID: %s", advisorID)
			}
			
			student.AdvisorID = &advisorID
		}

		// Create student
		_, err := s.studentRepo.Create(student)
		return err

	case "Dosen Wali":
		if req.LecturerID == nil || *req.LecturerID == "" {
			return errors.New("lecturerId is required for Dosen Wali role")
		}
		if req.Department == nil || *req.Department == "" {
			return errors.New("department is required for Dosen Wali role")
		}

		lecturer := models.Lecturer{
			ID:         uuid.New(),  // <- INI PENTING JUGA!
			UserID:     user.ID,
			LecturerID: *req.LecturerID,
			Department: *req.Department,
			CreatedAt:  time.Now(),
		}

		_, err := s.lecturerRepo.Create(lecturer)
		return err

	default:
		return nil
	}
}

// Delete godoc
// @Summary Delete user (soft delete)
// @Description Soft delete user by ID. Admin only.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{} "User deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid user ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 404 {object} map[string]interface{} "Not Found - User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id} [delete]
func (s *UserService) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Check if user exists (active only)
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check user",
			"details": err.Error(),
		})
	}
	if user == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found or already inactive",
		})
	}

	// Soft delete user
	if err := s.userRepo.SoftDelete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete user",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully (soft delete)",
	})
}

// Update godoc
// @Summary Update user
// @Description Update user information. Admin only.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body models.UpdateUserRequest true "Update data"
// @Success 200 {object} map[string]interface{} "User updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid data"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 404 {object} map[string]interface{} "Not Found - User not found"
// @Failure 409 {object} map[string]interface{} "Conflict - Email already used"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id} [put]
func (s *UserService) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if user exists (active only)
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check user",
			"details": err.Error(),
		})
	}
	if user == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found or inactive",
		})
	}

	// Validate email uniqueness if updating email
	if req.Email != nil {
		existingUser, _ := s.userRepo.GetByEmail(*req.Email)
		if existingUser != nil && existingUser.ID != id {
			return c.Status(409).JSON(fiber.Map{
				"error": "Email already used by another user",
			})
		}
	}

	// Update password if provided
	if req.Password != nil {
		hashedPassword, err := utils.HashPassword(*req.Password)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to hash password",
				"details": err.Error(),
			})
		}
		
		if err := s.userRepo.UpdatePassword(id, hashedPassword); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to update password",
				"details": err.Error(),
			})
		}
	}

	// Update other user data
	if err := s.userRepo.Update(id, &req); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update user",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User updated successfully",
	})
}

// UpdateRole godoc
// @Summary Update user role
// @Description Update user role. Admin only.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body map[string]interface{} true "Role update data" SchemaExample({"roleId": "7013c2d3-53dd-402a-81b2-0a8988acdc0a"})
// @Success 200 {object} map[string]interface{} "User role updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid user/role ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 404 {object} map[string]interface{} "Not Found - User/Role not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id}/role [put]
func (s *UserService) UpdateRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req struct {
		RoleID string `json:"roleId"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check user",
			"details": err.Error(),
		})
	}
	if user == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found or inactive",
		})
	}

	// Check if role exists
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check role",
			"details": err.Error(),
		})
	}
	if role == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	// Update role
	roleIDStr := roleID.String()
	updateReq := &models.UpdateUserRequest{
		RoleID: &roleIDStr,
	}
	
	if err := s.userRepo.Update(id, updateReq); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update role",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User role updated successfully",
	})
}

// SearchByName godoc
// @Summary Search users by name
// @Description Search users by full name. Admin only.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name query string true "Name to search"
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Items per page" minimum(1) maximum(100) default(10)
// @Success 200 {object} map[string]interface{} "Search results with pagination"
// @Failure 400 {object} map[string]interface{} "Bad Request - Name parameter required"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/search [get]
func (s *UserService) SearchByName(c *fiber.Ctx) error {
	name := c.Query("name", "")
	if name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name parameter is required",
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := s.userRepo.SearchByName(name, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to search users",
			"details": err.Error(),
		})
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"data": users,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}