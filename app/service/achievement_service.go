package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"UAS/app/models"
	"UAS/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementService struct {
	achievementRepo    repository.AchievementRepository
	achievementRefRepo repository.AchievementReferenceRepository
	studentRepo        repository.StudentRepository
	lecturerRepo       repository.LecturerRepository
	userRepo           repository.UserRepository
	roleRepo           repository.RoleRepository
}

func NewAchievementService(
	achievementRepo repository.AchievementRepository,
	achievementRefRepo repository.AchievementReferenceRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
) *AchievementService {
	return &AchievementService{
		achievementRepo:    achievementRepo,
		achievementRefRepo: achievementRefRepo,
		studentRepo:        studentRepo,
		lecturerRepo:       lecturerRepo,
		userRepo:           userRepo,
		roleRepo:           roleRepo,
	}
}

// ==================== HELPER ====================
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *AchievementService) GetAllAchievements(c *fiber.Ctx) error {
	ctx := context.Background()

	// Get user dari middleware
	userID := c.Locals("user_id").(uuid.UUID)
	user := c.Locals("user").(*models.User)

	// Get user role
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Get query parameters
	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	achievementType := c.Query("type", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var references []models.AchievementReference
	var total int

	// Role-based access
	switch userRole.Name {
	case "Admin":
		// Admin bisa lihat semua
		offset := (page - 1) * limit
		references, total, err = s.achievementRefRepo.GetAllReferences(status, limit, offset)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievements"})
		}

	case "Dosen Wali":
		// Dosen hanya bisa lihat mahasiswa bimbingannya
		lecturer, err := s.lecturerRepo.GetByUserID(userID)
		if err != nil || lecturer == nil {
			return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"})
		}
		references, err = s.achievementRefRepo.GetReferencesByAdvisor(lecturer.ID, status)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievements"})
		}
		total = len(references)

	case "Mahasiswa":
		// Mahasiswa hanya bisa lihat miliknya sendiri
		student, err := s.studentRepo.GetByUserID(userID)
		if err != nil || student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "User is not a student"})
		}
		references, err = s.achievementRefRepo.GetReferencesByStudentID(student.ID, status)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievements"})
		}
		total = len(references)

	default:
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// CEK JIKA TIDAK ADA DATA LANGSUNG RETURN
	if len(references) == 0 {
		return c.JSON(fiber.Map{
			"success": true,
			"data":    []interface{}{},
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total":       0,
				"total_pages": 0,
				"has_next":    false,
				"has_prev":    false,
			},
		})
	}

	// Get MongoDB IDs
	var mongoIDs []string
	for _, ref := range references {
		mongoIDs = append(mongoIDs, ref.MongoAchievementID)
	}

	// Get achievements dari MongoDB - TAMBAH CHECK ERROR KHUSUS
	var achievements []models.Achievement
	if len(mongoIDs) > 0 {
		achievements, err = s.achievementRepo.GetAchievementsByIDs(ctx, mongoIDs)
		if err != nil {
			// JANGAN return error, tapi log dan lanjut dengan array kosong
			fmt.Printf("Warning: Failed to get achievement details: %v\n", err)
			achievements = []models.Achievement{}
		}
	}

	// Combine data
	achievementMap := make(map[string]models.Achievement)
	for _, achievement := range achievements {
		achievementMap[achievement.ID.Hex()] = achievement
	}

	// Apply filters
	var results []fiber.Map
	for _, ref := range references {
		achievement, exists := achievementMap[ref.MongoAchievementID]
		if !exists {
			// Jika achievement tidak ditemukan di MongoDB, skip atau tampilkan data minimal
			student, _ := s.studentRepo.GetByID(ref.StudentID)
			var studentName, studentIDStr string
			if student != nil {
				studentUser, _ := s.userRepo.GetByID(student.UserID)
				if studentUser != nil {
					studentName = studentUser.FullName
				}
				studentIDStr = student.StudentID
			}

			// Data minimal jika achievement tidak ditemukan di MongoDB
			results = append(results, fiber.Map{
				"id":           ref.ID,
				"status":       ref.Status,
				"title":        "Achievement data not available",
				"type":         "unknown",
				"points":       0,
				"submitted_at": ref.SubmittedAt,
				"verified_at":  ref.VerifiedAt,
				"created_at":   ref.CreatedAt,
				"student": fiber.Map{
					"id":         student.ID,
					"name":       studentName,
					"student_id": studentIDStr,
				},
			})
			continue
		}

		// Filter by type
		if achievementType != "" && achievement.AchievementType != achievementType {
			continue
		}

		// Filter by search
		if search != "" {
			searchLower := strings.ToLower(search)
			titleMatch := strings.Contains(strings.ToLower(achievement.Title), searchLower)
			descMatch := strings.Contains(strings.ToLower(achievement.Description), searchLower)
			tagMatch := false
			for _, tag := range achievement.Tags {
				if strings.Contains(strings.ToLower(tag), searchLower) {
					tagMatch = true
					break
				}
			}

			if !titleMatch && !descMatch && !tagMatch {
				continue
			}
		}

		// Get student info
		student, _ := s.studentRepo.GetByID(ref.StudentID)
		var studentName, studentIDStr string
		if student != nil {
			studentUser, _ := s.userRepo.GetByID(student.UserID)
			if studentUser != nil {
				studentName = studentUser.FullName
			}
			studentIDStr = student.StudentID
		}

		results = append(results, fiber.Map{
			"id":           ref.ID,
			"status":       ref.Status,
			"title":        achievement.Title,
			"type":         achievement.AchievementType,
			"points":       achievement.Points,
			"submitted_at": ref.SubmittedAt,
			"verified_at":  ref.VerifiedAt,
			"created_at":   ref.CreatedAt,
			"student": fiber.Map{
				"id":         student.ID,
				"name":       studentName,
				"student_id": studentIDStr,
			},
		})
	}

	// Update total setelah filtering
	total = len(results)

	// Apply pagination
	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}
	if start >= total {
		start = 0
		end = 0
	}

	paginatedData := results[start:end]

	return c.JSON(fiber.Map{
		"success": true,
		"data":    paginatedData,
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

func (s *AchievementService) GetAchievementByID(c *fiber.Ctx) error {
	ctx := context.Background()

	// Parse ID sebagai UUID PostgreSQL
	refUUID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// 1. Get reference dari PostgreSQL
	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement"})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// 2. Get achievement dari MongoDB
	achievement, err := s.achievementRepo.GetAchievementByID(ctx, ref.MongoAchievementID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement details"})
	}
	if achievement == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement details not found"})
	}

	// 3. Validate user access
	userID := c.Locals("user_id").(uuid.UUID)
	user := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	canAccess := false
	switch userRole.Name {
	case "Admin":
		canAccess = true
	case "Mahasiswa":
		student, err := s.studentRepo.GetByUserID(userID)
		if err == nil && student != nil && student.ID == ref.StudentID {
			canAccess = true
		}
	case "Dosen Wali":
		lecturer, err := s.lecturerRepo.GetByUserID(userID)
		if err == nil && lecturer != nil {
			student, err := s.studentRepo.GetByID(ref.StudentID)
			if err == nil && student != nil && student.AdvisorID != nil && *student.AdvisorID == lecturer.ID {
				canAccess = true
			}
		}
	}

	if !canAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// 4. Get student info
	student, _ := s.studentRepo.GetByID(ref.StudentID)
	var studentInfo fiber.Map
	if student != nil {
		studentUser, _ := s.userRepo.GetByID(student.UserID)
		studentInfo = fiber.Map{
			"id":            student.ID,
			"student_id":    student.StudentID, 
			"program_study": student.ProgramStudy,
			"academic_year": student.AcademicYear,
		}
		if studentUser != nil {
			studentInfo["name"] = studentUser.FullName 
			studentInfo["username"] = studentUser.Username
			studentInfo["email"] = studentUser.Email
		}
	}

	// 5. Get verified by info jika ada
	var verifiedByInfo fiber.Map
	if ref.VerifiedBy != nil {
		verifiedUser, _ := s.userRepo.GetByID(*ref.VerifiedBy)
		if verifiedUser != nil {
			verifiedRole, _ := s.roleRepo.GetByID(verifiedUser.RoleID)
			roleName := ""
			if verifiedRole != nil {
				roleName = verifiedRole.Name
			}
			verifiedByInfo = fiber.Map{
				"id":   verifiedUser.ID,
				"name": verifiedUser.FullName,
				"role": roleName,
			}
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			// PostgreSQL ID (utama)
			"id": ref.ID,

			// Achievement data dari MongoDB
			"achievement_type": achievement.AchievementType,
			"title":            achievement.Title,
			"description":      achievement.Description,
			"points":           achievement.Points,
			"tags":             achievement.Tags,
			"details":          achievement.Details,
			"attachments":      achievement.Attachments,

			// Status info dari PostgreSQL
			"status":         ref.Status,
			"submitted_at":   ref.SubmittedAt,
			"verified_at":    ref.VerifiedAt,
			"verified_by":    verifiedByInfo,
			"rejection_note": ref.RejectionNote,

			// Student info
			"student":    studentInfo,
			"student_id": ref.StudentID,

			// Timestamps
			"created_at": ref.CreatedAt,
			"updated_at": ref.UpdatedAt,
		},
	})
}

func (s *AchievementService) CreateAchievement(c *fiber.Ctx) error {
	ctx := context.Background()

	// Get user info
	userID := c.Locals("user_id").(uuid.UUID)
	user := c.Locals("user").(*models.User)

	// Get user role
	role, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || role == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get user role",
		})
	}

	// Parse request body
	var req models.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if req.Title == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	if req.AchievementType == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Achievement type is required",
		})
	}

	// Validate achievement type
	validTypes := []string{"academic", "competition", "organization", "publication", "certification", "other"}
	if !contains(validTypes, req.AchievementType) {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid achievement type. Valid types: %v", validTypes),
		})
	}

	var studentID uuid.UUID
	var studentName string

	// Determine target student ID based on role
	switch role.Name {
	case "Mahasiswa":
		// Mahasiswa can only create achievements for themselves
		student, err := s.studentRepo.GetByUserID(userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to get student profile",
				"details": err.Error(),
			})
		}
		
		if student == nil {
			return c.Status(403).JSON(fiber.Map{
				"error": "User is not a student or student profile not found",
			})
		}
		
		studentID = student.ID
		
		// Get student name from user
		studentUser, _ := s.userRepo.GetByID(student.UserID)
		if studentUser != nil {
			studentName = studentUser.FullName
		}

	case "Admin":
		// Admin perlu menyertakan student_id di body
		var adminReq struct {
			models.CreateAchievementRequest
			StudentID *string `json:"student_id,omitempty"` // Changed to string for parsing
		}
		
		if err := c.BodyParser(&adminReq); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Invalid request body for admin",
				"details": err.Error(),
			})
		}
		
		if adminReq.StudentID == nil || *adminReq.StudentID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "student_id is required when creating achievement as admin",
			})
		}
		
		// Parse student_id string to UUID
		parsedStudentID, err := uuid.Parse(*adminReq.StudentID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid student_id format",
			})
		}
		
		student, err := s.studentRepo.GetByID(parsedStudentID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to get student",
				"details": err.Error(),
			})
		}
		
		if student == nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Student not found",
			})
		}
		
		studentID = parsedStudentID
		req = adminReq.CreateAchievementRequest
		
		// Get student name
		studentUser, _ := s.userRepo.GetByID(student.UserID)
		if studentUser != nil {
			studentName = studentUser.FullName
		}

	case "Dosen Wali":
		return c.Status(403).JSON(fiber.Map{
			"error": "Dosen Wali cannot create achievements",
		})

	default:
		return c.Status(403).JSON(fiber.Map{
			"error": "Unauthorized role: " + role.Name,
		})
	}

	// Initialize Attachments slice if nil
	if req.Attachments == nil {
		req.Attachments = []models.Attachment{}
	}
	
	// Initialize Tags slice if nil
	if req.Tags == nil {
		req.Tags = []string{}
	}

	// Create achievement object untuk MongoDB
	achievement := &models.Achievement{
		ID:              primitive.NewObjectID(),
		StudentID:       studentID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Attachments:     req.Attachments,
		Tags:            req.Tags,
		Points:          req.Points,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 1. Simpan ke MongoDB dulu
	mongoID, err := s.achievementRepo.CreateAchievement(ctx, achievement)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create achievement in MongoDB",
			"details": err.Error(),
		})
	}

	// 2. Buat reference di PostgreSQL dengan UUID sebagai ID API
	refID := uuid.New() // Ini ID yang akan dipakai di API
	ref := &models.AchievementReference{
		ID:                 refID,
		StudentID:          studentID,
		MongoAchievementID: mongoID,
		Status:             models.AchievementStatusDraft,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Create reference in PostgreSQL
	if err := s.achievementRefRepo.CreateReference(ref); err != nil {
		// Rollback: delete from MongoDB
		_ = s.achievementRepo.DeleteAchievement(ctx, mongoID)
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create achievement reference in PostgreSQL",
			"details": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Achievement created successfully",
		"data": fiber.Map{
			"id":                refID, 
			"mongo_id":          mongoID, 
			"student_id":        studentID,
			"student_name":      studentName,
			"achievement_type":  req.AchievementType,
			"title":             req.Title,
			"description":       req.Description,
			"status":            ref.Status,
			"points":            req.Points,
			"created_at":        ref.CreatedAt,
			"created_by":        userID,
			"created_by_name":   user.FullName,
		},
	})
}
// ==================== 4. UPDATE ACHIEVEMENT ====================
func (s *AchievementService) UpdateAchievement(c *fiber.Ctx) error {
	ctx := context.Background()

	// Parse ID sebagai UUID PostgreSQL
	refUUID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// 1. Get reference
	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// 2. Validate user access
	userID := c.Locals("user_id").(uuid.UUID)
	user := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Ownership check
	switch userRole.Name {
	case "Mahasiswa":
		student, err := s.studentRepo.GetByUserID(userID)
		if err != nil || student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "Student not found"})
		}
		if student.ID != ref.StudentID {
			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
		}
	case "Admin":
		// Admin bisa update semua
	default:
		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized role"})
	}

	// 3. Status check (hanya draft yang bisa diupdate)
	if ref.Status != models.AchievementStatusDraft {
		return c.Status(400).JSON(fiber.Map{
			"error": "Only draft achievements can be updated",
		})
	}

	// 4. Get achievement dari MongoDB
	achievement, err := s.achievementRepo.GetAchievementByID(ctx, ref.MongoAchievementID)
	if err != nil || achievement == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement details not found"})
	}

	// 5. Parse request body
	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 6. Apply updates
	if title, ok := req["title"].(string); ok && title != "" {
		achievement.Title = title
	}
	if description, ok := req["description"].(string); ok {
		achievement.Description = description
	}
	if points, ok := req["points"].(float64); ok {
		achievement.Points = int(points)
	}
	if tags, ok := req["tags"].([]interface{}); ok && len(tags) > 0 {
		var newTags []string
		for _, tag := range tags {
			if str, ok := tag.(string); ok {
				newTags = append(newTags, str)
			}
		}
		if len(newTags) > 0 {
			achievement.Tags = newTags
		}
	}

	// 7. Update details jika ada
	if details, ok := req["details"].(map[string]interface{}); ok {
		// Update competition fields
		if compName, ok := details["competition_name"].(string); ok && compName != "" {
			achievement.Details.CompetitionName = compName
		}
		if compLevel, ok := details["competition_level"].(string); ok && compLevel != "" {
			achievement.Details.CompetitionLevel = compLevel
		}
		if rank, ok := details["rank"].(float64); ok {
			achievement.Details.Rank = int(rank)
		}
		if medalType, ok := details["medal_type"].(string); ok && medalType != "" {
			achievement.Details.MedalType = medalType
		}

		// Update publication fields
		if pubType, ok := details["publication_type"].(string); ok && pubType != "" {
			achievement.Details.PublicationType = pubType
		}
		if pubTitle, ok := details["publication_title"].(string); ok && pubTitle != "" {
			achievement.Details.PublicationTitle = pubTitle
		}
		if authors, ok := details["authors"].([]interface{}); ok && len(authors) > 0 {
			var authorList []string
			for _, author := range authors {
				if str, ok := author.(string); ok {
					authorList = append(authorList, str)
				}
			}
			achievement.Details.Authors = authorList
		}
		if publisher, ok := details["publisher"].(string); ok && publisher != "" {
			achievement.Details.Publisher = publisher
		}
		if issn, ok := details["issn"].(string); ok && issn != "" {
			achievement.Details.ISSN = issn
		}
	}

	achievement.UpdatedAt = time.Now()

	// 8. Update di MongoDB
	if err := s.achievementRepo.UpdateAchievement(ctx, ref.MongoAchievementID, achievement); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update achievement"})
	}

	// 9. Update timestamp di PostgreSQL
	ref.UpdatedAt = time.Now()
	if err := s.achievementRefRepo.UpdateReference(ref); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update achievement reference"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement updated successfully",
		"data": fiber.Map{
			"id":         ref.ID,
			"status":     ref.Status,
			"updated_at": achievement.UpdatedAt,
		},
	})
}

func (s *AchievementService) DeleteAchievement(c *fiber.Ctx) error {
	// Parse ID sebagai UUID PostgreSQL
	refUUID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// 1. Get reference
	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get achievement",
			"details": err.Error(),
		})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// 2. Validate user access
	userID := c.Locals("user_id").(uuid.UUID)
	user := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Ownership check
	switch userRole.Name {
	case "Mahasiswa":
		student, err := s.studentRepo.GetByUserID(userID)
		if err != nil || student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "Student not found"})
		}
		if student.ID != ref.StudentID {
			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
		}
	case "Admin":
		// Admin bisa delete semua
	default:
		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized role"})
	}

	if ref.Status != models.AchievementStatusDraft {
		return c.Status(400).JSON(fiber.Map{
			"error": "Only draft achievements can be deleted",
		})
	}

	err = s.achievementRefRepo.SoftDelete(refUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete achievement",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement deleted successfully",
		"data": fiber.Map{
			"id":              ref.ID,
			"previous_status": ref.Status,
			"new_status":      models.AchievementStatusDeleted,
		},
	})
}

func (s *AchievementService) SubmitAchievement(c *fiber.Ctx) error {
	refUUID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// 1. Get reference
	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// 2. Validate user access
	userID := c.Locals("user_id").(uuid.UUID)
	user := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Status check
	if ref.Status != models.AchievementStatusDraft {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Only draft achievements can be submitted. Current: %s", ref.Status),
		})
	}

	// Jika Mahasiswa, cek ownership
	if userRole.Name == "Mahasiswa" {
		student, err := s.studentRepo.GetByUserID(userID)
		if err != nil || student == nil || student.ID != ref.StudentID {
			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
		}
	}
	// Admin tidak perlu validasi ownership

	// 3. Submit
	if err := s.achievementRefRepo.SubmitForVerification(refUUID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to submit achievement"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement submitted",
		"data": fiber.Map{
			"id":           ref.ID,
			"new_status":   models.AchievementStatusSubmitted,
			"submitted_at": time.Now(),
		},
	})
}

// ==================== 7. VERIFY ACHIEVEMENT ====================
func (s *AchievementService) VerifyAchievement(c *fiber.Ctx) error {
	// Parse ID sebagai UUID PostgreSQL
	refUUID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// 1. Get reference
	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// 2. Validate user access
	userID := c.Locals("user_id").(uuid.UUID)
	user := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Status check
	if ref.Status != models.AchievementStatusSubmitted {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Only submitted achievements can be verified. Current: %s", ref.Status),
		})
	}

	// Jika Dosen, cek apakah mahasiswa bimbingannya
	if userRole.Name == "Dosen Wali" {
		lecturer, err := s.lecturerRepo.GetByUserID(userID)
		if err != nil || lecturer == nil {
			return c.Status(403).JSON(fiber.Map{"error": "Lecturer not found"})
		}

		student, err := s.studentRepo.GetByID(ref.StudentID)
		if err != nil || student == nil || student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return c.Status(403).JSON(fiber.Map{
				"error": "You can only verify achievements of your advisees",
			})
		}
	}

	// 3. Verify
	if err := s.achievementRefRepo.VerifyAchievement(refUUID, userID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to verify achievement"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement verified",
		"data": fiber.Map{
			"id":          ref.ID,
			"new_status":  models.AchievementStatusVerified,
			"verified_by": userID,
			"verified_at": time.Now(),
		},
	})
}

func (s *AchievementService) RejectAchievement(c *fiber.Ctx) error {
	// Parse ID sebagai UUID PostgreSQL
	refUUID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// 1. Get reference
	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// 2. Validate user access
	userID := c.Locals("user_id").(uuid.UUID)
	user := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Parse rejection note
	var req struct {
		RejectionNote string `json:"rejection_note"`
	}
	if err := c.BodyParser(&req); err != nil || req.RejectionNote == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Rejection note is required"})
	}

	// Status check
	if ref.Status != models.AchievementStatusSubmitted {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Only submitted achievements can be rejected. Current: %s", ref.Status),
		})
	}

	// Jika Dosen, cek advisor
	if userRole.Name == "Dosen Wali" {
		lecturer, err := s.lecturerRepo.GetByUserID(userID)
		if err != nil || lecturer == nil {
			return c.Status(403).JSON(fiber.Map{"error": "Lecturer not found"})
		}

		student, err := s.studentRepo.GetByID(ref.StudentID)
		if err != nil || student == nil || student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return c.Status(403).JSON(fiber.Map{
				"error": "You can only reject achievements of your advisees",
			})
		}
	}

	// 3. Reject
	if err := s.achievementRefRepo.RejectAchievement(refUUID, userID, req.RejectionNote); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to reject achievement"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement rejected",
		"data": fiber.Map{
			"id":             ref.ID,
			"new_status":     models.AchievementStatusRejected,
			"rejection_note": req.RejectionNote,
			"rejected_by":    userID,
			"rejected_at":    time.Now(),
		},
	})
}

// ==================== 9. GET ACHIEVEMENT HISTORY ====================
func (s *AchievementService) GetAchievementHistory(c *fiber.Ctx) error {
	// Parse ID sebagai UUID PostgreSQL
	refUUID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// 1. Get reference
	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// 2. Validate user access
	userID := c.Locals("user_id").(uuid.UUID)
	user := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	canAccess := false
	switch userRole.Name {
	case "Admin":
		canAccess = true
	case "Mahasiswa":
		student, err := s.studentRepo.GetByUserID(userID)
		if err == nil && student != nil && student.ID == ref.StudentID {
			canAccess = true
		}
	case "Dosen Wali":
		lecturer, err := s.lecturerRepo.GetByUserID(userID)
		if err == nil && lecturer != nil {
			student, err := s.studentRepo.GetByID(ref.StudentID)
			if err == nil && student != nil && student.AdvisorID != nil && *student.AdvisorID == lecturer.ID {
				canAccess = true
			}
		}
	}

	if !canAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// 3. Build history
	history := []fiber.Map{
		{
			"status":      models.AchievementStatusDraft,
			"changed_at":  ref.CreatedAt,
			"changed_by":  nil,
			"description": "Achievement created",
		},
	}

	if ref.SubmittedAt != nil {
		history = append(history, fiber.Map{
			"status":      models.AchievementStatusSubmitted,
			"changed_at":  ref.SubmittedAt,
			"changed_by":  nil,
			"description": "Submitted for verification",
		})
	}

	if ref.Status == models.AchievementStatusVerified && ref.VerifiedAt != nil && ref.VerifiedBy != nil {
		verifierUser, _ := s.userRepo.GetByID(*ref.VerifiedBy)
		var verifierName string
		if verifierUser != nil {
			verifierName = verifierUser.FullName
		}

		history = append(history, fiber.Map{
			"status":     models.AchievementStatusVerified,
			"changed_at": ref.VerifiedAt,
			"changed_by": fiber.Map{
				"id":   ref.VerifiedBy,
				"name": verifierName,
			},
			"description": "Verified by advisor",
		})
	}

	if ref.Status == models.AchievementStatusRejected && ref.VerifiedAt != nil && ref.VerifiedBy != nil {
		rejectorUser, _ := s.userRepo.GetByID(*ref.VerifiedBy)
		var rejectorName string
		if rejectorUser != nil {
			rejectorName = rejectorUser.FullName
		}

		history = append(history, fiber.Map{
			"status":     models.AchievementStatusRejected,
			"changed_at": ref.VerifiedAt,
			"changed_by": fiber.Map{
				"id":   ref.VerifiedBy,
				"name": rejectorName,
			},
			"description": fmt.Sprintf("Rejected: %s", *ref.RejectionNote),
		})
	}

	// 4. Get achievement title
	ctx := context.Background()
	achievement, _ := s.achievementRepo.GetAchievementByID(ctx, ref.MongoAchievementID)
	var title string
	if achievement != nil {
		title = achievement.Title
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"achievement_id": ref.ID,
			"title":          title,
			"current_status": ref.Status,
			"history":        history,
		},
	})
}

// ==================== 10. UPLOAD ATTACHMENT ====================
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	ctx := context.Background()

	// Parse ID sebagai UUID PostgreSQL
	refUUID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// 1. Get reference
	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// 2. Validate user access
	userID := c.Locals("user_id").(uuid.UUID)
	user := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	canUpload := false
	switch userRole.Name {
	case "Admin":
		canUpload = true
	case "Mahasiswa":
		student, err := s.studentRepo.GetByUserID(userID)
		if err == nil && student != nil && student.ID == ref.StudentID {
			canUpload = true
		}
	}

	if !canUpload {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// 3. Status check (hanya draft yang bisa upload attachment)
	if ref.Status != models.AchievementStatusDraft {
		return c.Status(400).JSON(fiber.Map{
			"error": "Only draft achievements can have attachments uploaded",
		})
	}

	// 4. Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File is required"})
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return c.Status(400).JSON(fiber.Map{"error": "File too large (max 10MB)"})
	}

	// Validate file type
	allowedTypes := []string{
		"application/pdf",
		"image/jpeg", "image/jpg", "image/png",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}

	fileType := file.Header.Get("Content-Type")
	isAllowed := false
	for _, allowed := range allowedTypes {
		if fileType == allowed {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return c.Status(400).JSON(fiber.Map{
			"error": "File type not allowed. Allowed: PDF, JPEG, PNG, DOC, DOCX",
		})
	}

	// 5. Save file
	uploadDir := "./uploads/achievements"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	fileExt := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%d%s",
		strings.TrimSuffix(file.Filename, fileExt),
		time.Now().Unix(),
		fileExt,
	)

	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
	}

	// 6. Create attachment object
	attachment := models.Attachment{
		FileName:   file.Filename,
		FileURL:    "/uploads/achievements/" + filename,
		FileType:   fileType,
		UploadedAt: time.Now(),
	}

	// 7. Save to MongoDB
	err = s.achievementRepo.AddAttachment(ctx, ref.MongoAchievementID, attachment)
	if err != nil {
		os.Remove(filePath)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save attachment"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Attachment uploaded successfully",
		"data": fiber.Map{
			"attachment":     attachment,
			"achievement_id": ref.ID,
			"uploaded_at":    attachment.UploadedAt,
		},
	})
}
