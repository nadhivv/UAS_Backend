package service

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"time"

// 	"UAS/app/models"
// 	"UAS/app/repository"

// 	"go.mongodb.org/mongo-driver/bson/primitive"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/google/uuid"
// )

// type AchievementService struct {
// 	achievementRepo    repository.AchievementRepository
// 	achievementRefRepo repository.AchievementReferenceRepository
// 	studentRepo        repository.StudentRepository
// 	lecturerRepo       repository.LecturerRepository
// 	userRepo           repository.UserRepository
// 	roleRepo           repository.RoleRepository
// }

// func NewAchievementService(
// 	achievementRepo repository.AchievementRepository,
// 	achievementRefRepo repository.AchievementReferenceRepository,
// 	studentRepo repository.StudentRepository,
// 	lecturerRepo repository.LecturerRepository,
// 	userRepo repository.UserRepository,
// 	roleRepo repository.RoleRepository,
// ) *AchievementService {
// 	return &AchievementService{
// 		achievementRepo:    achievementRepo,
// 		achievementRefRepo: achievementRefRepo,
// 		studentRepo:        studentRepo,
// 		lecturerRepo:       lecturerRepo,
// 		userRepo:           userRepo,
// 		roleRepo:           roleRepo,
// 	}
// }

// func (s *AchievementService) CreateAchievement(c *fiber.Ctx) error {
// 	ctx := context.Background()

// 	// Get user info from JWT
// 	userIDStr, err := s.getUserIDFromContext(c)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{
// 			"error": "Unauthorized",
// 		})
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
// 	}

// 	user, err := s.userRepo.GetByID(userID)
// 	if err != nil || user == nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
// 	}

// 	// Get user's role
// 	role, err := s.roleRepo.GetByID(user.RoleID)
// 	if err != nil || role == nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Failed to get user role",
// 		})
// 	}

// 	// Parse request body
// 	var req models.CreateAchievementRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error":   "Invalid request body",
// 			"details": err.Error(),
// 		})
// 	}

// 	// Validate achievement type
// 	validTypes := []string{"academic", "competition", "organization", "publication", "certification", "other"}
// 	if !contains(validTypes, req.AchievementType) {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": fmt.Sprintf("Invalid achievement type. Valid types: %v", validTypes),
// 		})
// 	}

// 	var studentID uuid.UUID

// 	// Determine target student ID based on role
// 	switch role.Name {
// 	case "Mahasiswa":
// 		// Mahasiswa can only create achievements for themselves
// 		student, err := s.studentRepo.GetByUserID(userID)
// 		if err != nil || student == nil {
// 			return c.Status(403).JSON(fiber.Map{
// 				"error": "User is not a student or student profile not found",
// 			})
// 		}
// 		studentID = student.ID

// 	case "Admin":
// 		// Admin perlu menyertakan student_id di body
// 		var adminReq struct {
// 			models.CreateAchievementRequest
// 			StudentID *uuid.UUID `json:"student_id,omitempty"`
// 		}
		
// 		if err := c.BodyParser(&adminReq); err != nil {
// 			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
// 		}
		
// 		if adminReq.StudentID == nil || *adminReq.StudentID == uuid.Nil {
// 			return c.Status(400).JSON(fiber.Map{
// 				"error": "student_id is required when creating achievement as admin",
// 			})
// 		}

// 		student, err := s.studentRepo.GetByID(adminReq.StudentID)
// 		if err != nil || student == nil {
// 			return c.Status(404).JSON(fiber.Map{
// 				"error": "Student not found",
// 			})
// 		}
// 		studentID = *adminReq.StudentID
// 		req = adminReq.CreateAchievementRequest

// 	case "Dosen Wali":
// 		return c.Status(403).JSON(fiber.Map{
// 			"error": "Dosen Wali cannot create achievements",
// 		})

// 	default:
// 		return c.Status(403).JSON(fiber.Map{
// 			"error": "Unauthorized role",
// 		})
// 	}

// 	// Create achievement object
// 	achievement := &models.Achievement{
// 		StudentID:       studentID,
// 		AchievementType: req.AchievementType,
// 		Title:           req.Title,
// 		Description:     req.Description,
// 		Details:         req.Details,
// 		Attachments:     req.Attachments,
// 		Tags:            req.Tags,
// 		Points:          req.Points,
// 		CreatedAt:       time.Now(),
// 		UpdatedAt:       time.Now(),
// 	}

// 	// Save to MongoDB
// 	mongoID, err := s.achievementRepo.CreateAchievement(ctx, achievement)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error":   "Failed to create achievement",
// 			"details": err.Error(),
// 		})
// 	}

// 	// Create reference in PostgreSQL
// 	ref := &models.AchievementReference{
// 		StudentID:          studentID,
// 		MongoAchievementID: mongoID,
// 		Status:             models.AchievementStatusDraft,
// 		CreatedAt:          time.Now(),
// 		UpdatedAt:          time.Now(),
// 	}

// 	if err := s.achievementRefRepo.CreateReference(ref); err != nil {
// 		// Rollback: delete from MongoDB
// 		s.achievementRepo.DeleteAchievement(ctx, mongoID)
// 		return c.Status(500).JSON(fiber.Map{
// 			"error":   "Failed to create achievement reference",
// 			"details": err.Error(),
// 		})
// 	}

// 	return c.Status(201).JSON(fiber.Map{
// 		"success": true,
// 		"message": "Achievement created successfully",
// 		"data": fiber.Map{
// 			"id":               mongoID,
// 			"reference_id":     ref.ID,
// 			"student_id":       studentID,
// 			"achievement_type": req.AchievementType,
// 			"title":            req.Title,
// 			"description":      req.Description,
// 			"status":           ref.Status,
// 			"points":           req.Points,
// 			"created_at":       ref.CreatedAt,
// 			"created_by":       userID,
// 		},
// 	})
// }

// func (s *AchievementService) GetAchievementByID(c *fiber.Ctx) error {
// 	ctx := context.Background()
	
// 	achievementID := c.Params("id")
	
// 	// Parse sebagai UUID (reference ID)
// 	refUUID, err := uuid.Parse(achievementID)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
// 	}

// 	// Get reference dari PostgreSQL
// 	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement"})
// 	}
// 	if ref == nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
// 	}

// 	// Get achievement dari MongoDB
// 	achievement, err := s.achievementRepo.GetAchievementByID(ctx, ref.MongoAchievementID)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement details"})
// 	}
// 	if achievement == nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Achievement details not found"})
// 	}

// 	// Get user info untuk validasi access
// 	userIDStr, err := s.getUserIDFromContext(c)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
// 	}

// 	user, err := s.userRepo.GetByID(userID)
// 	if err != nil || user == nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
// 	}

// 	userRole, err := s.roleRepo.GetByID(user.RoleID)
// 	if err != nil || userRole == nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
// 	}

// 	canAccess := false
	
// 	switch userRole.Name {
// 	case "Admin":
// 		// Admin bisa akses semua
// 		canAccess = true
		
// 	case "Mahasiswa":
// 		// Mahasiswa hanya bisa akses miliknya sendiri
// 		student, err := s.studentRepo.GetByUserID(userID)
// 		if err == nil && student != nil && student.ID == ref.StudentID {
// 			canAccess = true
// 		}
		
// 	case "Dosen Wali":
// 		// Dosen hanya bisa akses mahasiswa bimbingannya
// 		lecturer, err := s.lecturerRepo.GetByUserID(userID)
// 		if err == nil && lecturer != nil {
// 			// Cek apakah student ini adalah bimbingan dosen
// 			student, err := s.studentRepo.GetByID(ref.StudentID)
// 			if err == nil && student != nil && student.AdvisorID != nil && *student.AdvisorID == lecturer.ID {
// 				canAccess = true
// 			}
// 		}
// 	}
	
// 	if !canAccess {
// 		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
// 	}

// 	// Get student info
// 	student, _ := s.studentRepo.GetByID(ref.StudentID)
// 	var studentInfo fiber.Map
// 	if student != nil {
// 		studentUser, _ := s.userRepo.GetByID(student.UserID)
// 		studentInfo = fiber.Map{
// 			"id":            student.ID,
// 			"nim":           student.NIM,
// 			"name":          student.Name,
// 			"program_study": student.ProgramStudy,
// 			"academic_year": student.AcademicYear,
// 		}
// 		if studentUser != nil {
// 			studentInfo["user_name"] = studentUser.Username
// 			studentInfo["email"] = studentUser.Email
// 		}
// 	}

// 	var verifiedByInfo fiber.Map
// 	if ref.VerifiedBy != nil {
// 		verifiedUser, _ := s.userRepo.GetByID(*ref.VerifiedBy)
// 		if verifiedUser != nil {
// 			verifiedByInfo = fiber.Map{
// 				"id":   verifiedUser.ID,
// 				"name": verifiedUser.FullName,
// 				"role": userRole.Name,
// 			}
// 		}
// 	}

// 	// Convert MongoDB ObjectID to string
// 	mongoIDStr := achievement.ID.Hex()

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"data": fiber.Map{
// 			// IDs
// 			"id":                 ref.ID,
// 			"mongo_id":           mongoIDStr,
			
// 			// Achievement data
// 			"achievement_type":   achievement.AchievementType,
// 			"title":              achievement.Title,
// 			"description":        achievement.Description,
// 			"points":             achievement.Points,
// 			"tags":               achievement.Tags,
// 			"details":            achievement.Details,
// 			"attachments":        achievement.Attachments,
			
// 			// Status info
// 			"status":            ref.Status,
// 			"submitted_at":      ref.SubmittedAt,
// 			"verified_at":       ref.VerifiedAt,
// 			"verified_by":       verifiedByInfo,
// 			"rejection_note":    ref.RejectionNote,
			
// 			// Student info
// 			"student":          studentInfo,
// 			"student_id":       ref.StudentID,
			
// 			// Timestamps
// 			"created_at":       ref.CreatedAt,
// 			"updated_at":       ref.UpdatedAt,
// 		},
// 	})
// }

// // GetMyAchievements - Mahasiswa melihat prestasi sendiri
// func (s *AchievementService) GetMyAchievements(c *fiber.Ctx) error {
// 	ctx := context.Background()

// 	userIDStr, err := s.getUserIDFromContext(c)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
// 	}

// 	// Get student ID
// 	student, err := s.studentRepo.GetByUserID(userID)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to get student profile"})
// 	}
// 	if student == nil {
// 		return c.Status(403).JSON(fiber.Map{"error": "User is not a student"})
// 	}

// 	// Get query parameters
// 	status := c.Query("status", "")
// 	page := c.QueryInt("page", 1)
// 	limit := c.QueryInt("limit", 10)
// 	search := c.Query("search", "")

// 	if page < 1 {
// 		page = 1
// 	}
// 	if limit < 1 || limit > 100 {
// 		limit = 10
// 	}

// 	// Get references from PostgreSQL
// 	references, err := s.achievementRefRepo.GetReferencesByStudentID(student.ID, status)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error":   "Failed to get achievements",
// 			"details": err.Error(),
// 		})
// 	}

// 	if len(references) == 0 {
// 		return c.JSON(fiber.Map{
// 			"success": true,
// 			"data": []interface{}{},
// 			"pagination": fiber.Map{
// 				"page":  page,
// 				"limit": limit,
// 				"total": 0,
// 			},
// 		})
// 	}

// 	// Get MongoDB IDs
// 	var mongoIDs []string
// 	for _, ref := range references {
// 		mongoIDs = append(mongoIDs, ref.MongoAchievementID)
// 	}

// 	// Get achievements from MongoDB
// 	achievements, err := s.achievementRepo.GetAchievementsByIDs(ctx, mongoIDs)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error":   "Failed to get achievement details",
// 			"details": err.Error(),
// 		})
// 	}

// 	// Combine data and apply search filter
// 	var filteredAchievements []fiber.Map
// 	achievementMap := make(map[string]models.Achievement)
// 	for _, achievement := range achievements {
// 		achievementMap[achievement.ID.Hex()] = achievement
// 	}

// 	for _, ref := range references {
// 		achievement, exists := achievementMap[ref.MongoAchievementID]
// 		if !exists {
// 			continue
// 		}

// 		// Apply search filter
// 		if search != "" {
// 			searchLower := strings.ToLower(search)
// 			titleMatch := strings.Contains(strings.ToLower(achievement.Title), searchLower)
// 			descMatch := strings.Contains(strings.ToLower(achievement.Description), searchLower)
// 			tagMatch := false
// 			for _, tag := range achievement.Tags {
// 				if strings.Contains(strings.ToLower(tag), searchLower) {
// 					tagMatch = true
// 					break
// 				}
// 			}

// 			if !titleMatch && !descMatch && !tagMatch {
// 				continue
// 			}
// 		}

// 		filteredAchievements = append(filteredAchievements, fiber.Map{
// 			"id":           ref.ID,
// 			"status":       ref.Status,
// 			"title":        achievement.Title,
// 			"type":         achievement.AchievementType,
// 			"points":       achievement.Points,
// 			"submitted_at": ref.SubmittedAt,
// 			"verified_at":  ref.VerifiedAt,
// 			"created_at":   ref.CreatedAt,
// 		})
// 	}

// 	// Apply pagination
// 	total := len(filteredAchievements)
// 	totalPages := (total + limit - 1) / limit
// 	hasNext := page < totalPages
// 	hasPrev := page > 1

// 	start := (page - 1) * limit
// 	end := start + limit
// 	if end > total {
// 		end = total
// 	}
// 	if start >= total {
// 		start = 0
// 		end = 0
// 	}

// 	paginatedData := filteredAchievements[start:end]

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"data": paginatedData,
// 		"pagination": fiber.Map{
// 			"page":        page,
// 			"limit":       limit,
// 			"total":       total,
// 			"total_pages": totalPages,
// 			"has_next":    hasNext,
// 			"has_prev":    hasPrev,
// 		},
// 	})
// }

// func (s *AchievementService) UpdateAchievement(c *fiber.Ctx) error {
// 	ctx := context.Background()
	
// 	achievementID := c.Params("id")
// 	refUUID, err := uuid.Parse(achievementID)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
// 	}

// 	// Get reference
// 	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
// 	if err != nil || ref == nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
// 	}

// 	// Get user info
// 	userIDStr, err := s.getUserIDFromContext(c)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
// 	}

// 	user, err := s.userRepo.GetByID(userID)
// 	if err != nil || user == nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
// 	}

// 	// Get user role
// 	userRole, err := s.roleRepo.GetByID(user.RoleID)
// 	if err != nil || userRole == nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
// 	}

// 	// Ownership check berdasarkan role
// 	switch userRole.Name {
// 	case "Mahasiswa":
// 		// Mahasiswa hanya bisa update miliknya sendiri
// 		student, err := s.studentRepo.GetByUserID(userID)
// 		if err != nil || student == nil {
// 			return c.Status(403).JSON(fiber.Map{"error": "Student not found"})
// 		}
		
// 		if student.ID != ref.StudentID {
// 			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
// 		}
// 	case "Admin":
// 		// Admin bisa update semua
// 		// Tidak perlu validasi ownership
// 	default:
// 		// Dosen & lainnya tidak boleh update
// 		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized role"})
// 	}

// 	// Status check (hanya draft yang bisa diupdate)
// 	if ref.Status != models.AchievementStatusDraft {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": "Only draft achievements can be updated",
// 			"current_status": ref.Status,
// 		})
// 	}

// 	// Get achievement dari MongoDB
// 	achievement, err := s.achievementRepo.GetAchievementByID(ctx, ref.MongoAchievementID)
// 	if err != nil || achievement == nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Achievement details not found"})
// 	}

// 	// Parse request body sebagai map
// 	var req map[string]interface{}
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
// 	}

// 	// Apply updates
// 	if title, ok := req["title"].(string); ok && title != "" {
// 		achievement.Title = title
// 	}
// 	if description, ok := req["description"].(string); ok {
// 		achievement.Description = description
// 	}
// 	if points, ok := req["points"].(float64); ok {
// 		achievement.Points = int(points)
// 	}
// 	if tags, ok := req["tags"].([]interface{}); ok && len(tags) > 0 {
// 		var newTags []string
// 		for _, tag := range tags {
// 			if str, ok := tag.(string); ok {
// 				newTags = append(newTags, str)
// 			}
// 		}
// 		if len(newTags) > 0 {
// 			achievement.Tags = newTags
// 		}
// 	}

// 	// Update details jika ada
// 	if details, ok := req["details"].(map[string]interface{}); ok {
// 		// Update competition fields
// 		if compName, ok := details["competition_name"].(string); ok && compName != "" {
// 			achievement.Details.CompetitionName = compName
// 		}
// 		if compLevel, ok := details["competition_level"].(string); ok && compLevel != "" {
// 			achievement.Details.CompetitionLevel = compLevel
// 		}
// 		if rank, ok := details["rank"].(float64); ok {
// 			achievement.Details.Rank = int(rank)
// 		}
// 		if medalType, ok := details["medal_type"].(string); ok && medalType != "" {
// 			achievement.Details.MedalType = medalType
// 		}
		
// 		// Update publication fields
// 		if pubType, ok := details["publication_type"].(string); ok && pubType != "" {
// 			achievement.Details.PublicationType = pubType
// 		}
// 		if pubTitle, ok := details["publication_title"].(string); ok && pubTitle != "" {
// 			achievement.Details.PublicationTitle = pubTitle
// 		}
// 		if authors, ok := details["authors"].([]interface{}); ok && len(authors) > 0 {
// 			var authorList []string
// 			for _, author := range authors {
// 				if str, ok := author.(string); ok {
// 					authorList = append(authorList, str)
// 				}
// 			}
// 			achievement.Details.Authors = authorList
// 		}
// 		if publisher, ok := details["publisher"].(string); ok && publisher != "" {
// 			achievement.Details.Publisher = publisher
// 		}
// 		if issn, ok := details["issn"].(string); ok && issn != "" {
// 			achievement.Details.ISSN = issn
// 		}
		
// 		// Update organization fields
// 		if orgName, ok := details["organization_name"].(string); ok && orgName != "" {
// 			achievement.Details.OrganizationName = orgName
// 		}
// 		if position, ok := details["position"].(string); ok && position != "" {
// 			achievement.Details.Position = position
// 		}
		
// 		// Update period
// 		if period, ok := details["period"].(map[string]interface{}); ok {
// 			var periodObj models.Period
// 			if startStr, ok := period["start"].(string); ok && startStr != "" {
// 				if startTime, err := time.Parse(time.RFC3339, startStr); err == nil {
// 					periodObj.Start = startTime
// 				}
// 			}
// 			if endStr, ok := period["end"].(string); ok && endStr != "" {
// 				if endTime, err := time.Parse(time.RFC3339, endStr); err == nil {
// 					periodObj.End = endTime
// 				}
// 			}
// 			achievement.Details.Period = &periodObj
// 		}
		
// 		// Update certification fields
// 		if certName, ok := details["certification_name"].(string); ok && certName != "" {
// 			achievement.Details.CertificationName = certName
// 		}
// 		if issuedBy, ok := details["issued_by"].(string); ok && issuedBy != "" {
// 			achievement.Details.IssuedBy = issuedBy
// 		}
// 		if certNumber, ok := details["certification_number"].(string); ok && certNumber != "" {
// 			achievement.Details.CertificationNumber = certNumber
// 		}
// 		if validUntilStr, ok := details["valid_until"].(string); ok && validUntilStr != "" {
// 			if validUntil, err := time.Parse(time.RFC3339, validUntilStr); err == nil {
// 				achievement.Details.ValidUntil = &validUntil
// 			}
// 		}
		
// 		// Update event fields
// 		if eventDateStr, ok := details["event_date"].(string); ok && eventDateStr != "" {
// 			if eventDate, err := time.Parse(time.RFC3339, eventDateStr); err == nil {
// 				achievement.Details.EventDate = &eventDate
// 			}
// 		}
// 		if location, ok := details["location"].(string); ok && location != "" {
// 			achievement.Details.Location = location
// 		}
// 		if organizer, ok := details["organizer"].(string); ok && organizer != "" {
// 			achievement.Details.Organizer = organizer
// 		}
// 		if score, ok := details["score"].(float64); ok {
// 			achievement.Details.Score = int(score)
// 		}
		
// 		// Update custom fields
// 		if customFields, ok := details["custom_fields"].(map[string]interface{}); ok {
// 			achievement.Details.CustomFields = customFields
// 		}
// 	}

// 	achievement.UpdatedAt = time.Now()

// 	// Update di MongoDB
// 	if err := s.achievementRepo.UpdateAchievement(ctx, ref.MongoAchievementID, achievement); err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to update achievement"})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Achievement updated successfully",
// 		"data": fiber.Map{
// 			"id":         ref.ID,
// 			"mongo_id":   ref.MongoAchievementID,
// 			"status":     ref.Status,
// 			"updated_at": achievement.UpdatedAt,
// 		},
// 	})
// }

// func (s *AchievementService) DeleteAchievement(c *fiber.Ctx) error {
// 	ctx := context.Background()
	
// 	achievementID := c.Params("id")
// 	refUUID, err := uuid.Parse(achievementID)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
// 	}

// 	// Get reference
// 	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
// 	if err != nil || ref == nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
// 	}

// 	// Get user info
// 	userIDStr, err := s.getUserIDFromContext(c)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
// 	}

// 	user, err := s.userRepo.GetByID(userID)
// 	if err != nil || user == nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
// 	}

// 	userRole, err := s.roleRepo.GetByID(user.RoleID)
// 	if err != nil || userRole == nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
// 	}

// 	// Ownership check berdasarkan role
// 	switch userRole.Name {
// 	case "Mahasiswa":
// 		// Mahasiswa hanya bisa delete miliknya sendiri
// 		student, err := s.studentRepo.GetByUserID(userID)
// 		if err != nil || student == nil {
// 			return c.Status(403).JSON(fiber.Map{"error": "Student not found"})
// 		}
		
// 		if student.ID != ref.StudentID {
// 			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
// 		}
// 	case "Admin":
// 		// Admin bisa delete semua
// 	default:
// 		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized role"})
// 	}

// 	// Status check (hanya draft yang bisa di-delete)
// 	if ref.Status != models.AchievementStatusDraft {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": "Only draft achievements can be deleted",
// 			"current_status": ref.Status,
// 		})
// 	}

// 	// Delete from MongoDB
// 	err = s.achievementRepo.DeleteAchievement(ctx, ref.MongoAchievementID)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Failed to delete achievement from MongoDB",
// 		})
// 	}

// 	// Delete reference from PostgreSQL
// 	err = s.achievementRefRepo.DeleteReference(refUUID)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Failed to delete achievement reference",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Achievement deleted successfully",
// 		"data": fiber.Map{
// 			"id":              ref.ID,
// 			"mongo_id":        ref.MongoAchievementID,
// 			"previous_status": ref.Status,
// 			"deleted_at":      time.Now(),
// 		},
// 	})
// }

// func (s *AchievementService) SubmitAchievement(c *fiber.Ctx) error {
// 	achievementID := c.Params("id")
// 	refUUID, err := uuid.Parse(achievementID)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
// 	}

// 	// Get user info
// 	userIDStr, err := s.getUserIDFromContext(c)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
// 	}

// 	user, err := s.userRepo.GetByID(userID)
// 	if err != nil || user == nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
// 	}

// 	// Get reference
// 	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
// 	if err != nil || ref == nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
// 	}

// 	// Cek status
// 	if ref.Status != models.AchievementStatusDraft {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": fmt.Sprintf("Only draft achievements can be submitted. Current: %s", ref.Status),
// 		})
// 	}

// 	// Jika Mahasiswa, cek ownership
// 	userRole, err := s.roleRepo.GetByID(user.RoleID)
// 	if err != nil || userRole == nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
// 	}

// 	if userRole.Name == "Mahasiswa" {
// 		student, err := s.studentRepo.GetByUserID(userID)
// 		if err != nil || student == nil || student.ID != ref.StudentID {
// 			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
// 		}
// 	}
// 	// Admin tidak perlu validasi ownership

// 	// Submit
// 	if err := s.achievementRefRepo.SubmitForVerification(refUUID); err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to submit achievement"})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Achievement submitted",
// 		"data": fiber.Map{
// 			"id":          ref.ID,
// 			"new_status":  models.AchievementStatusSubmitted,
// 			"submitted_at": time.Now(),
// 		},
// 	})
// }

// // GetAdviseeAchievements - Dosen melihat prestasi mahasiswa bimbingan
// func (s *AchievementService) GetAdviseeAchievements(c *fiber.Ctx) error {
// 	ctx := context.Background()

// 	userIDStr, err := s.getUserIDFromContext(c)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
// 	}

// 	// Get lecturer ID
// 	lecturer, err := s.lecturerRepo.GetByUserID(userID)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to get lecturer profile"})
// 	}
// 	if lecturer == nil {
// 		return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"})
// 	}

// 	// Get query parameters
// 	status := c.Query("status", "")
// 	page := c.QueryInt("page", 1)
// 	limit := c.QueryInt("limit", 10)

// 	if page < 1 {
// 		page = 1
// 	}
// 	if limit < 1 || limit > 100 {
// 		limit = 10
// 	}

// 	// Get references from PostgreSQL
// 	references, err := s.achievementRefRepo.GetReferencesByAdvisor(lecturer.ID, status)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error":   "Failed to get achievements",
// 			"details": err.Error(),
// 		})
// 	}

// 	// Get achievement details from MongoDB
// 	var achievements []fiber.Map
// 	for _, ref := range references {
// 		achievement, err := s.achievementRepo.GetAchievementByID(ctx, ref.MongoAchievementID)
// 		if err != nil {
// 			continue
// 		}

// 		// Get student info
// 		student, _ := s.studentRepo.GetByID(ref.StudentID)
// 		var studentName string
// 		if student != nil {
// 			studentUser, _ := s.userRepo.GetByID(student.UserID)
// 			if studentUser != nil {
// 				studentName = studentUser.FullName
// 			}
// 		}

// 		achievements = append(achievements, fiber.Map{
// 			"id":           ref.ID,
// 			"status":       ref.Status,
// 			"title":        achievement.Title,
// 			"type":         achievement.AchievementType,
// 			"points":       achievement.Points,
// 			"submitted_at": ref.SubmittedAt,
// 			"verified_at":  ref.VerifiedAt,
// 			"created_at":   ref.CreatedAt,
// 			"student": fiber.Map{
// 				"id":   student.ID,
// 				"name": studentName,
// 				"nim":  student.NIM,
// 			},
// 		})
// 	}

// 	// Apply pagination
// 	total := len(achievements)
// 	totalPages := (total + limit - 1) / limit
// 	hasNext := page < totalPages
// 	hasPrev := page > 1

// 	start := (page - 1) * limit
// 	end := start + limit
// 	if end > total {
// 		end = total
// 	}
// 	if start >= total {
// 		start = 0
// 		end = 0
// 	}

// 	paginatedData := achievements[start:end]

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"data": paginatedData,
// 		"pagination": fiber.Map{
// 			"page":        page,
// 			"limit":       limit,
// 			"total":       total,
// 			"total_pages": totalPages,
// 			"has_next":    hasNext,
// 			"has_prev":    hasPrev,
// 		},
// 	})
// }

// func (s *AchievementService) VerifyAchievement(c *fiber.Ctx) error {
// 	achievementID := c.Params("id")
// 	refUUID, err := uuid.Parse(achievementID)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
// 	}

// 	// Get user info
// 	userIDStr, err := s.getUserIDFromContext(c)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
// 	}

// 	user, err := s.userRepo.GetByID(userID)
// 	if err != nil || user == nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
// 	}

// 	// Get user role
// 	userRole, err := s.roleRepo.GetByID(user.RoleID)
// 	if err != nil || userRole == nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
// 	}

// 	// Get reference
// 	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
// 	if err != nil || ref == nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
// 	}

// 	// Cek status
// 	if ref.Status != models.AchievementStatusSubmitted {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": fmt.Sprintf("Only submitted achievements can be verified. Current: %s", ref.Status),
// 		})
// 	}

// 	// Jika user adalah Dosen, cek apakah mahasiswa bimbingannya
// 	if userRole.Name == "Dosen Wali" {
// 		lecturer, err := s.lecturerRepo.GetByUserID(userID)
// 		if err != nil || lecturer == nil {
// 			return c.Status(403).JSON(fiber.Map{"error": "Lecturer not found"})
// 		}
		
// 		student, err := s.studentRepo.GetByID(ref.StudentID)
// 		if err != nil || student == nil || student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
// 			return c.Status(403).JSON(fiber.Map{
// 				"error": "You can only verify achievements of your advisees",
// 			})
// 		}
// 	}

// 	// Verify
// 	if err := s.achievementRefRepo.VerifyAchievement(refUUID, userID); err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to verify achievement"})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Achievement verified",
// 		"data": fiber.Map{
// 			"id":         ref.ID,
// 			"new_status": models.AchievementStatusVerified,
// 			"verified_by": userID,
// 			"verified_at": time.Now(),
// 		},
// 	})
// }

// func (s *AchievementService) RejectAchievement(c *fiber.Ctx) error {
// 	achievementID := c.Params("id")
// 	refUUID, err := uuid.Parse(achievementID)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
// 	}

// 	// Get user info
// 	userIDStr, err := s.getUserIDFromContext(c)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
// 	}

// 	user, err := s.userRepo.GetByID(userID)
// 	if err != nil || user == nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
// 	}

// 	// Get user role
// 	userRole, err := s.roleRepo.GetByID(user.RoleID)
// 	if err != nil || userRole == nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
// 	}

// 	// Parse rejection note
// 	var req struct {
// 		RejectionNote string `json:"rejection_note"`
// 	}
// 	if err := c.BodyParser(&req); err != nil || req.RejectionNote == "" {
// 		return c.Status(400).JSON(fiber.Map{"error": "Rejection note is required"})
// 	}

// 	// Get reference
// 	ref, err := s.achievementRefRepo.GetReferenceByID(refUUID)
// 	if err != nil || ref == nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
// 	}

// 	// Cek status
// 	if ref.Status != models.AchievementStatusSubmitted {
// 		return c.Status(400).JSON(fiber.Map{
// 			"error": fmt.Sprintf("Only submitted achievements can be rejected. Current: %s", ref.Status),
// 		})
// 	}

// 	// Jika Dosen, cek advisor
// 	if userRole.Name == "Dosen Wali" {
// 		lecturer, err := s.lecturerRepo.GetByUserID(userID)
// 		if err != nil || lecturer == nil {
// 			return c.Status(403).JSON(fiber.Map{"error": "Lecturer not found"})
// 		}
		
// 		student, err := s.studentRepo.GetByID(ref.StudentID)
// 		if err != nil || student == nil || student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
// 			return c.Status(403).JSON(fiber.Map{
// 				"error": "You can only reject achievements of your advisees",
// 			})
// 		}
// 	}

// 	// Reject
// 	if err := s.achievementRefRepo.RejectAchievement(refUUID, userID, req.RejectionNote); err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to reject achievement"})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"message": "Achievement rejected",
// 		"data": fiber.Map{
// 			"id":             ref.ID,
// 			"new_status":     models.AchievementStatusRejected,
// 			"rejection_note": req.RejectionNote,
// 			"rejected_by":    userID,
// 			"rejected_at":    time.Now(),
// 		},
// 	})
// }

// func (s *AchievementService) GetAllAchievements(c *fiber.Ctx) error {
// 	ctx := context.Background()

// 	// Get user info
// 	userIDStr, err := s.getUserIDFromContext(c)
// 	if err != nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
// 	}

// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
// 	}

// 	user, err := s.userRepo.GetByID(userID)
// 	if err != nil || user == nil {
// 		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
// 	}

// 	// Get user role
// 	userRole, err := s.roleRepo.GetByID(user.RoleID)
// 	if err != nil || userRole == nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
// 	}

// 	// Get query parameters
// 	status := c.Query("status", "")
// 	page := c.QueryInt("page", 1)
// 	limit := c.QueryInt("limit", 10)

// 	if page < 1 {
// 		page = 1
// 	}
// 	if limit < 1 || limit > 100 {
// 		limit = 10
// 	}

// 	var references []models.AchievementReference
// 	var total int

// 	// Role-based access
// 	switch userRole.Name {
// 	case "Admin":
// 		// Admin bisa lihat semua
// 		offset := (page - 1) * limit
// 		references, total, err = s.achievementRefRepo.GetAllReferences(status, limit, offset)
// 		if err != nil {
// 			return c.Status(500).JSON(fiber.Map{
// 				"error":   "Failed to get achievements",
// 				"details": err.Error(),
// 			})
// 		}
// 	case "Dosen Wali":
// 		// Dosen hanya bisa lihat mahasiswa bimbingannya
// 		lecturer, err := s.lecturerRepo.GetByUserID(userID)
// 		if err != nil || lecturer == nil {
// 			return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"})
// 		}
// 		references, err = s.achievementRefRepo.GetReferencesByAdvisor(lecturer.ID, status)
// 		if err != nil {
// 			return c.Status(500).JSON(fiber.Map{
// 				"error":   "Failed to get achievements",
// 				"details": err.Error(),
// 			})
// 		}
// 		total = len(references)
// 	default:
// 		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
// 	}

// 	// Get MongoDB IDs
// 	var mongoIDs []string
// 	for _, ref := range references {
// 		mongoIDs = append(mongoIDs, ref.MongoAchievementID)
// 	}

// 	// Get achievements from MongoDB
// 	achievements, err := s.achievementRepo.GetAchievementsByIDs(ctx, mongoIDs)
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error":   "Failed to get achievement details",
// 			"details": err.Error(),
// 		})
// 	}

// 	// Combine data
// 	achievementMap := make(map[string]models.Achievement)
// 	for _, achievement := range achievements {
// 		achievementMap[achievement.ID.Hex()] = achievement
// 	}

// 	var result []fiber.Map
// 	for _, ref := range references {
// 		achievement, exists := achievementMap[ref.MongoAchievementID]
// 		if !exists {
// 			continue
// 		}

// 		// Get student info
// 		student, _ := s.studentRepo.GetByID(ref.StudentID)
// 		var studentName, studentNIM string
// 		if student != nil {
// 			studentUser, _ := s.userRepo.GetByID(student.UserID)
// 			if studentUser != nil {
// 				studentName = studentUser.FullName
// 			}
// 			studentNIM = student.NIM
// 		}

// 		result = append(result, fiber.Map{
// 			"id":           ref.ID,
// 			"status":       ref.Status,
// 			"title":        achievement.Title,
// 			"type":         achievement.AchievementType,
// 			"points":       achievement.Points,
// 			"submitted_at": ref.SubmittedAt,
// 			"verified_at":  ref.VerifiedAt,
// 			"created_at":   ref.CreatedAt,
// 			"student": fiber.Map{
// 				"id":   student.ID,
// 				"name": studentName,
// 				"nim":  studentNIM,
// 			},
// 		})
// 	}

// 	// Apply pagination for Dosen Wali
// 	if userRole.Name == "Dosen Wali" {
// 		totalPages := (total + limit - 1) / limit
// 		hasNext := page < totalPages
// 		hasPrev := page > 1

// 		start := (page - 1) * limit
// 		end := start + limit
// 		if end > total {
// 			end = total
// 		}
// 		if start >= total {
// 			start = 0
// 			end = 0
// 		}

// 		result = result[start:end]

// 		return c.JSON(fiber.Map{
// 			"success": true,
// 			"data": result,
// 			"pagination": fiber.Map{
// 				"page":        page,
// 				"limit":       limit,
// 				"total":       total,
// 				"total_pages": totalPages,
// 				"has_next":    hasNext,
// 				"has_prev":    hasPrev,
// 			},
// 		})
// 	}

// 	// For Admin with built-in pagination
// 	totalPages := (total + limit - 1) / limit
// 	hasNext := page < totalPages
// 	hasPrev := page > 1

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"data": result,
// 		"pagination": fiber.Map{
// 			"page":        page,
// 			"limit":       limit,
// 			"total":       total,
// 			"total_pages": totalPages,
// 			"has_next":    hasNext,
// 			"has_prev":    hasPrev,
// 		},
// 	})
// }

// // Helper untuk mendapatkan user ID dari context Fiber
// func (s *AchievementService) getUserIDFromContext(c *fiber.Ctx) (string, error) {
// 	userIDVal := c.Locals("user_id")
// 	if userIDVal == nil {
// 		return "", fmt.Errorf("user_id not found in context")
// 	}

// 	switch v := userIDVal.(type) {
// 	case string:
// 		return v, nil
// 	case uuid.UUID:
// 		return v.(), nil
// 	case []byte:
// 		return string(v), nil
// 	default:
// 		str := fmt.Sprintf("%v", v)
// 		// Try to parse as UUID
// 		if parsed, err := uuid.Parse(str); err == nil {
// 			return parsed.(), nil
// 		}
// 		return str, nil
// 	}
// }

// // Helper function to check if a string exists in a slice
// func contains(slice []string, item string) bool {
// 	for _, s := range slice {
// 		if s == item {
// 			return true
// 		}
// 	}
// 	return false
// }