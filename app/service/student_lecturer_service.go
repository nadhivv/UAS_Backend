package service

import (
	"context"

	"UAS/app/models"
	"UAS/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StudentLecturerService struct {
	studentRepo        repository.StudentRepository
	lecturerRepo       repository.LecturerRepository
	userRepo           repository.UserRepository
	roleRepo           repository.RoleRepository
	achievementRepo    repository.AchievementRepository
	achievementRefRepo repository.AchievementReferenceRepository
}

func NewStudentLecturerService(
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	achievementRepo repository.AchievementRepository,
	achievementRefRepo repository.AchievementReferenceRepository,
) *StudentLecturerService {
	return &StudentLecturerService{
		studentRepo:        studentRepo,
		lecturerRepo:       lecturerRepo,
		userRepo:           userRepo,
		roleRepo:           roleRepo,
		achievementRepo:    achievementRepo,
		achievementRefRepo: achievementRefRepo,
	}
}


// 1. GET /api/v1/students
// GetAllStudents godoc
// @Summary Get all students
// @Description Get list of all students with user details and advisor information. Admin only.
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of students"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /students [get]
func (s *StudentLecturerService) GetAllStudents(c *fiber.Ctx) error {
	// Get all students dari repository
	students, err := s.studentRepo.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get students",
			"details": err.Error(),
		})
	}

	// Enrich dengan data user
	var enrichedStudents []models.StudentResponse
	for _, student := range students {
		user, err := s.userRepo.GetByID(student.UserID)
		if err != nil {
			continue // Skip jika user tidak ditemukan
		}

		// Get advisor info jika ada
		var advisorName string
		if student.AdvisorID != nil && *student.AdvisorID != uuid.Nil {
			lecturer, err := s.lecturerRepo.GetByID(*student.AdvisorID)
			if err == nil && lecturer != nil {
				advisorUser, _ := s.userRepo.GetByID(lecturer.UserID)
				if advisorUser != nil {
					advisorName = advisorUser.FullName
				}
			}
		}

		response := models.StudentResponse{
			ID:            student.ID,
			UserID:        student.UserID,
			StudentID:     student.StudentID,
			FullName:      user.FullName,
			Email:         user.Email,
			Username:      user.Username,
			ProgramStudy:  student.ProgramStudy,
			AcademicYear:  student.AcademicYear,
			AdvisorID:     student.AdvisorID,
			AdvisorName:   advisorName,
			CreatedAt:     student.CreatedAt,
		}
		enrichedStudents = append(enrichedStudents, response)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    enrichedStudents,
		"total":   len(enrichedStudents),
	})
}

// 2. GET /api/v1/students/:id
// GetStudentByID godoc
// @Summary Get student by ID
// @Description Get student details by ID. Admin: any student, Dosen Wali: only advisees, Mahasiswa: only self
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID (UUID)"
// @Success 200 {object} map[string]interface{} "Student details"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid student ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Access denied"
// @Failure 404 {object} map[string]interface{} "Not Found - Student not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /students/{id} [get]
func (s *StudentLecturerService) GetStudentByID(c *fiber.Ctx) error {
	studentIDStr := c.Params("id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID format"})
	}

	// Get student
	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get student",
			"details": err.Error(),
		})
	}
	if student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	// Get user details
	user, err := s.userRepo.GetByID(student.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get user details",
			"details": err.Error(),
		})
	}
	if user == nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found for this student"})
	}

	// Get advisor info jika ada
	var advisorName string
	var advisorUserDetails *models.User
	if student.AdvisorID != nil && *student.AdvisorID != uuid.Nil {
		lecturer, err := s.lecturerRepo.GetByID(*student.AdvisorID)
		if err == nil && lecturer != nil {
			advisorUserDetails, _ = s.userRepo.GetByID(lecturer.UserID)
			if advisorUserDetails != nil {
				advisorName = advisorUserDetails.FullName
			}
		}
	}

	response := models.StudentResponse{
		ID:            student.ID,
		UserID:        student.UserID,
		StudentID:     student.StudentID,
		FullName:      user.FullName,
		Email:         user.Email,
		Username:      user.Username,
		ProgramStudy:  student.ProgramStudy,
		AcademicYear:  student.AcademicYear,
		AdvisorID:     student.AdvisorID,
		AdvisorName:   advisorName,
		CreatedAt:     student.CreatedAt,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetStudentAchievements godoc
// @Summary Get student achievements
// @Description Get achievements for specific student. Admin: any student, Dosen Wali: only advisees, Mahasiswa: only self
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID (UUID)"
// @Param status query string false "Filter by status" Enums(draft, submitted, verified, rejected)
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Items per page" minimum(1) maximum(100) default(10)
// @Success 200 {object} map[string]interface{} "Student achievements list"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid student ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Access denied"
// @Failure 404 {object} map[string]interface{} "Not Found - Student not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /students/{id}/achievements [get]
func (s *StudentLecturerService) GetStudentAchievements(c *fiber.Ctx) error {
	studentIDStr := c.Params("id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID format"})
	}

	// Check if student exists
	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get student",
			"details": err.Error(),
		})
	}
	if student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	// Get query parameters
	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Get achievement references
	refs, total, err := s.achievementRefRepo.GetAllReferences(status, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get achievements",
			"details": err.Error(),
		})
	}

	// Filter by student ID dan get achievement details
	var achievements []fiber.Map
	for _, ref := range refs {
		if ref.StudentID != studentID {
			continue
		}

		ctx := context.Background()
		// Get MongoDB achievement details
		achievement, err := s.achievementRepo.GetAchievementByID(ctx, ref.MongoAchievementID)
		if err != nil {
			// Skip jika error, tapi log mungkin diperlukan
			continue
		}

		// Get verified by user info jika ada
		var verifiedByName string
		if ref.VerifiedBy != nil {
			verifiedUser, _ := s.userRepo.GetByID(*ref.VerifiedBy)
			if verifiedUser != nil {
				verifiedByName = verifiedUser.FullName
			}
		}

		// Format attachments jika ada
		var attachments []fiber.Map
		if achievement != nil && len(achievement.Attachments) > 0 {
			for _, att := range achievement.Attachments {
				attachments = append(attachments, fiber.Map{
					"file_name":   att.FileName,
					"file_url":    att.FileURL,
					"file_type":   att.FileType,
					"uploaded_at": att.UploadedAt,
				})
			}
		}

		achievements = append(achievements, fiber.Map{
			"id":              ref.ID,
			"mongo_id":        ref.MongoAchievementID,
			"title":           achievement.Title,
			"description":     achievement.Description,
			"achievement_type": achievement.AchievementType,
			"status":          ref.Status,
			"points":          achievement.Points,
			"submitted_at":    ref.SubmittedAt,
			"verified_at":     ref.VerifiedAt,
			"verified_by": fiber.Map{
				"id":   ref.VerifiedBy,
				"name": verifiedByName,
			},
			"rejection_note": ref.RejectionNote,
			"details":        achievement.Details,
			"tags":           achievement.Tags,
			"attachments":    attachments,
			"created_at":     achievement.CreatedAt,
			"updated_at":     achievement.UpdatedAt,
		})
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"student_id":    studentID,
			"achievements":  achievements,
			"total":         len(achievements),
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
			},
		},
	})
}

// UpdateStudentAdvisor godoc
// @Summary Update student advisor
// @Description Assign or remove advisor for student. Admin only.
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID (UUID)"
// @Param request body map[string]interface{} true "Advisor data" SchemaExample({"advisor_id": "123e4567-e89b-12d3-a456-426614174000"})
// @Success 200 {object} map[string]interface{} "Advisor updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid student/advisor ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 404 {object} map[string]interface{} "Not Found - Student/Lecturer not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /students/{id}/advisor [put]
func (s *StudentLecturerService) UpdateStudentAdvisor(c *fiber.Ctx) error {
	// Get student ID from params
	studentIDStr := c.Params("id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID"})
	}

	// Parse request body
	var req struct {
		AdvisorID *string `json:"advisor_id,omitempty"` // null untuk remove advisor
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check if student exists
	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get student"})
	}
	if student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	var advisorID *uuid.UUID
	var advisorName string

	// Handle advisor assignment/removal
	if req.AdvisorID == nil || *req.AdvisorID == "" {
		// Remove advisor
		if err := s.studentRepo.RemoveAdvisor(studentID); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to remove advisor",
				"details": err.Error(),
			})
		}
	} else {
		// Assign new advisor
		parsedAdvisorID, err := uuid.Parse(*req.AdvisorID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid advisor ID"})
		}
		advisorID = &parsedAdvisorID

		// Check if advisor exists and is actually a lecturer
		lecturer, err := s.lecturerRepo.GetByID(parsedAdvisorID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get advisor"})
		}
		if lecturer == nil {
			return c.Status(404).JSON(fiber.Map{"error": "Lecturer not found"})
		}

		// Get advisor user info for response
		advisorUser, _ := s.userRepo.GetByID(lecturer.UserID)
		if advisorUser != nil {
			advisorName = advisorUser.FullName
		}

		// Update advisor
		if err := s.studentRepo.UpdateAdvisor(studentID, advisorID); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to update advisor",
				"details": err.Error(),
			})
		}
	}

	// Get updated student info
	updatedStudent, _ := s.studentRepo.GetByID(studentID)
	studentUser, _ := s.userRepo.GetByID(updatedStudent.UserID)

	// Prepare response data
	advisorResponse := fiber.Map{}
	if advisorID != nil {
		advisorResponse["id"] = advisorID
		advisorResponse["name"] = advisorName
	} else {
		advisorResponse["id"] = nil
		advisorResponse["name"] = ""
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Student advisor updated successfully",
		"data": fiber.Map{
			"student": fiber.Map{
				"id":            updatedStudent.ID,
				"student_id":    updatedStudent.StudentID,
				"name":          studentUser.FullName,
				"program_study": updatedStudent.ProgramStudy,
				"academic_year": updatedStudent.AcademicYear,
			},
			"advisor": advisorResponse,
			"action": func() string {
				if req.AdvisorID == nil || *req.AdvisorID == "" {
					return "removed"
				}
				return "assigned"
			}(),
		},
	})
}

// GetAllLecturers godoc
// @Summary Get all lecturers
// @Description Get list of all lecturers with user details and advisees count. Admin only.
// @Tags Lecturers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Items per page" minimum(1) maximum(100) default(10)
// @Param search query string false "Search by lecturer name"
// @Param department query string false "Filter by department"
// @Success 200 {object} map[string]interface{} "List of lecturers"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin access required"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /lecturers [get]
func (s *StudentLecturerService) GetAllLecturers(c *fiber.Ctx) error {
	// Get query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	department := c.Query("department", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var lecturers []models.LecturerResponse
	var total int
	var err error

	// Apply filters
	if search != "" {
		lecturers, total, err = s.lecturerRepo.SearchByName(search, page, limit)
	} else if department != "" {
		rawLecturers, count, err := s.lecturerRepo.GetByDepartment(department, page, limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to get lecturers by department",
				"details": err.Error(),
			})
		}

		// Convert to LecturerResponse
		for _, rawLecturer := range rawLecturers {
			user, _ := s.userRepo.GetByID(rawLecturer.UserID)
			if user != nil {
				lecturers = append(lecturers, models.LecturerResponse{
					ID:         rawLecturer.ID,
					UserID:     rawLecturer.UserID,
					FullName:   user.FullName,
					Username:   user.Username,
					Email:      user.Email,
					LecturerID: rawLecturer.LecturerID,
					Department: rawLecturer.Department,
					CreatedAt:  rawLecturer.CreatedAt,
				})
			}
		}
		total = count
	} else {
		lecturers, total, err = s.lecturerRepo.GetWithUserDetails(page, limit)
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get lecturers",
			"details": err.Error(),
		})
	}

	// Add advisees count to each lecturer
	var lecturersWithCount []fiber.Map
	for _, lecturer := range lecturers {
		adviseesCount, _ := s.lecturerRepo.GetAdviseesCount(lecturer.ID)

		lecturerMap := fiber.Map{
			"id":            lecturer.ID,
			"user_id":       lecturer.UserID,
			"full_name":     lecturer.FullName,
			"email":         lecturer.Email,
			"username":      lecturer.Username,
			"lecturer_id":   lecturer.LecturerID,
			"department":    lecturer.Department,
			"created_at":    lecturer.CreatedAt,
			"advisees_count": adviseesCount,
		}
		lecturersWithCount = append(lecturersWithCount, lecturerMap)
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"success": true,
		"data":    lecturersWithCount,
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

// GetLecturerAdvisees godoc
// @Summary Get lecturer advisees
// @Description Get list of students supervised by specific lecturer. Admin: any lecturer, Dosen Wali: only self
// @Tags Lecturers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lecturer ID (UUID)"
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Items per page" minimum(1) maximum(100) default(10)
// @Success 200 {object} map[string]interface{} "List of advisees with achievement stats"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid lecturer ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Access denied"
// @Failure 404 {object} map[string]interface{} "Not Found - Lecturer not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /lecturers/{id}/advisees [get]
func (s *StudentLecturerService) GetLecturerAdvisees(c *fiber.Ctx) error {
	lecturerIDStr := c.Params("id")
	lecturerID, err := uuid.Parse(lecturerIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid lecturer ID format"})
	}

	// Check if lecturer exists
	lecturer, err := s.lecturerRepo.GetByID(lecturerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get lecturer",
			"details": err.Error(),
		})
	}
	if lecturer == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Lecturer not found"})
	}

	// Get query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get advisees
	students, total, err := s.lecturerRepo.GetAdvisees(lecturerID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get advisees",
			"details": err.Error(),
		})
	}

	// Enrich students with user details and achievement stats
	var advisees []fiber.Map
	for _, student := range students {
		user, _ := s.userRepo.GetByID(student.UserID)
		if user == nil {
			continue
		}

		// Get achievement statistics
		achievementRefs, _ := s.achievementRefRepo.GetReferencesByStudentID(student.ID, "")
		verifiedCount := 0
		for _, ref := range achievementRefs {
			if ref.Status == "verified" {
				verifiedCount++
			}
		}

		advisees = append(advisees, fiber.Map{
			"id":            student.ID,
			"user_id":       student.UserID,
			"student_id":    student.StudentID,
			"full_name":     user.FullName,
			"email":         user.Email,
			"program_study": student.ProgramStudy,
			"academic_year": student.AcademicYear,
			"created_at":    student.CreatedAt,
			"achievement_stats": fiber.Map{
				"total":    len(achievementRefs),
				"verified": verifiedCount,
			},
		})
	}

	// Get lecturer user info
	lecturerUser, _ := s.userRepo.GetByID(lecturer.UserID)

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"lecturer": fiber.Map{
				"id":          lecturer.ID,
				"lecturer_id": lecturer.LecturerID,
				"name":        lecturerUser.FullName,
				"department":  lecturer.Department,
			},
			"advisees":       advisees,
			"total_students": total,
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
			},
		},
	})
}