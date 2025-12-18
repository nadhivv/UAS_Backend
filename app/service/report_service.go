package service

import (
	"context"
	"time"

	"UAS/app/models"
	"UAS/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ReportService struct {
	reportRepo  repository.ReportRepository
	userRepo    repository.UserRepository
	studentRepo repository.StudentRepository
	lecturerRepo repository.LecturerRepository
	roleRepo    repository.RoleRepository
}

func NewReportService(
	reportRepo repository.ReportRepository,
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	roleRepo repository.RoleRepository,
) *ReportService {
	return &ReportService{
		reportRepo:  reportRepo,
		userRepo:    userRepo,
		studentRepo: studentRepo,
		lecturerRepo: lecturerRepo,
		roleRepo:    roleRepo,
	}
}

// GetStatistics godoc
// @Summary Get achievement statistics
// @Description Get achievement statistics based on user role. Admin: all statistics, Dosen Wali: advisee's statistics, Mahasiswa: own statistics
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (format: YYYY-MM-DD)" Example(2024-01-01)
// @Param end_date query string false "End date (format: YYYY-MM-DD)" Example(2024-12-31)
// @Success 200 {object} map[string]interface{} "Statistics data"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Access denied"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/statistics [get]
func (s *ReportService) GetStatistics(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	role, err := s.roleRepo.GetByID(currentUser.RoleID)
	if err != nil || role == nil {
		return c.Status(403).JSON(fiber.Map{"error": "invalid role"})
	}

	var scope string
	var actorID uuid.UUID

	switch role.Name {
	case "Admin":
		scope = "all"
		actorID = currentUser.ID
	case "Dosen Wali":
		lecturer, err := s.lecturerRepo.GetByUserID(currentUser.ID)
		if err != nil || lecturer == nil {
			return c.Status(403).JSON(fiber.Map{"error": "lecturer not found"})
		}
		scope = "lecturer"
		actorID = lecturer.ID
	case "Mahasiswa":
		student, err := s.studentRepo.GetByUserID(currentUser.ID)
		if err != nil || student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "student not found"})
		}
		scope = "student"
		actorID = student.ID
	default:
		return c.Status(403).JSON(fiber.Map{"error": "access denied"})
	}

	var startDate, endDate *time.Time
	if v := c.Query("start_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			startDate = &t
		}
	}
	if v := c.Query("end_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			endDate = &t
		}
	}

	ctx := c.UserContext()
	if ctx == nil {
		ctx = context.Background()
	}

	stats, err := s.reportRepo.GetStatistics(ctx, actorID, scope, startDate, endDate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// GetStudentReport godoc
// @Summary Get student achievement report
// @Description Get detailed achievement report for specific student. Admin: all students, Dosen Wali: only advisees, Mahasiswa: only self
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID (UUID)"
// @Param start_date query string false "Start date (format: YYYY-MM-DD)" Example(2024-01-01)
// @Param end_date query string false "End date (format: YYYY-MM-DD)" Example(2024-12-31)
// @Success 200 {object} map[string]interface{} "Student report data"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid student ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Access denied"
// @Failure 404 {object} map[string]interface{} "Not Found - Student not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/students/{id} [get]
func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
	studentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid student id"})
	}

	currentUser, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	role, err := s.roleRepo.GetByID(currentUser.RoleID)
	if err != nil || role == nil {
		return c.Status(403).JSON(fiber.Map{"error": "invalid role"})
	}

	authorized := false

	switch role.Name {
	case "Admin":
		authorized = true
	case "Dosen Wali":
		lecturer, _ := s.lecturerRepo.GetByUserID(currentUser.ID)
		student, _ := s.studentRepo.GetByID(studentID)
		if lecturer != nil && student != nil && student.AdvisorID != nil {
			authorized = *student.AdvisorID == lecturer.ID
		}
	case "Mahasiswa":
		student, _ := s.studentRepo.GetByID(studentID)
		if student != nil {
			authorized = student.UserID == currentUser.ID
		}
	}

	if !authorized {
		return c.Status(403).JSON(fiber.Map{"error": "access denied"})
	}

	var startDate, endDate *time.Time
	if v := c.Query("start_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			startDate = &t
		}
	}
	if v := c.Query("end_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			endDate = &t
		}
	}

	ctx := c.UserContext()
	if ctx == nil {
		ctx = context.Background()
	}

	stats, err := s.reportRepo.GetStatistics(ctx, studentID, "student", startDate, endDate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"student_id": studentID,
			"statistics": stats,
		},
	})
}
